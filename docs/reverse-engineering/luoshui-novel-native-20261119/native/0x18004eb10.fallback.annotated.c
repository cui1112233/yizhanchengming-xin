/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x18004eb10 */

longlong * FUN_18004eb10(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  int *piVar3;
  longlong *plVar4;
  int iVar5;
  longlong lVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  code *pcVar11;
  code *pcVar12;
  undefined8 uVar13;
  longlong *plVar14;
  longlong *local_res18;
  longlong local_res20;
  longlong *local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  undefined8 local_60;
  longlong *local_58;
  undefined8 local_50;
  longlong *local_48;
  longlong *local_40;
  
  plVar4 = DAT_1803ae750;
  plVar1 = (longlong *)param_3[3];
  plVar2 = (longlong *)*param_3;
  plVar14 = (longlong *)param_3[1];
  plVar9 = (longlong *)param_3[2];
  local_res20 = DAT_180391c90;
  local_98 = (longlong *)0x0;
  local_90 = (longlong *)0x0;
  local_88 = (longlong *)0x0;
  local_80 = (longlong *)0x0;
  local_res18 = (longlong *)0x0;
  local_a8 = plVar14;
  local_78 = plVar9;
  if (DAT_1803ae750 == (longlong *)0x0) {
LAB_18004eba3:
    DAT_1803ae750 = (longlong *)FUN_18036cd10(param_1,DAT_1803aec50,DAT_1803ca880,0x48);
  }
  else if ((1 < *DAT_1803ae750) || (DAT_1803ae750[2] != 0)) {
    if ((DAT_1803ae750 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803ae750 && (*DAT_1803ae750 = *DAT_1803ae750 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_18004eba3;
  }
  plVar4 = DAT_1803ae750;
  FUN_180001d00(param_1,DAT_1803ae750);
  lVar6 = FUN_180049c60(param_1);
  if (lVar6 == 0) {
/* novels.agents.reviser_agent#84="extract_tagged_block" */
    FUN_180373340(param_1,&local_res20,DAT_1803aea40);
    uVar13 = 0xb0;
    lVar6 = local_res20;
    plVar10 = (longlong *)0x0;
  }
  else {
    uVar13 = 0xb0;
/* novels.agents.reviser_agent#85="REVISED_CONTENT" */
    local_60 = DAT_1803aea48;
    *(undefined4 *)(plVar4 + 5) = 0xb0;
    local_68 = plVar2;
    plVar7 = (longlong *)FUN_180377750(param_1,lVar6,&local_68);
    if (plVar7 != (longlong *)0x0) {
      local_98 = plVar7;
      lVar6 = FUN_180049c60(param_1);
      if (lVar6 == 0) {
/* novels.agents.reviser_agent#84="extract_tagged_block" */
        FUN_180373340(param_1,&local_res20,DAT_1803aea40);
        uVar13 = 0xb1;
        lVar6 = local_res20;
        plVar10 = local_res18;
        goto LAB_18004f35e;
      }
      uVar13 = 0xb1;
/* novels.agents.reviser_agent#86="PATCHES" */
      local_50 = DAT_1803aea50;
      *(undefined4 *)(plVar4 + 5) = 0xb1;
      local_58 = plVar2;
      plVar8 = (longlong *)FUN_180377750(param_1,lVar6,&local_58);
      local_70 = plVar8;
      if (plVar8 != (longlong *)0x0) {
        local_90 = plVar8;
/* novels.agents.reviser_agent#87="spot-fix" */
        iVar5 = FUN_1803836d0(plVar9,DAT_1803aea58);
        if (iVar5 == -1) {
          uVar13 = 0xb3;
        }
        else {
          plVar10 = plVar14;
          if (iVar5 == 1) {
LAB_18004f1c9:
            lVar6 = FUN_180049d70(param_1);
            if (lVar6 == 0) {
/* novels.agents.reviser_agent#89="parse_spot_fix_patches" */
              FUN_180373340(param_1,&local_res20,DAT_1803aea68);
              uVar13 = 0xb4;
              lVar6 = local_res20;
              plVar10 = local_res18;
              goto LAB_18004f35e;
            }
            iVar5 = FUN_180001510(plVar8);
            uVar13 = 0xb4;
            if (iVar5 != -1) {
              *(undefined4 *)(plVar4 + 5) = 0xb4;
              plVar9 = plVar2;
              if (iVar5 == 1) {
                plVar9 = plVar8;
              }
              plVar9 = (longlong *)FUN_180376770(param_1,lVar6,plVar9);
              if (plVar9 != (longlong *)0x0) {
                local_88 = plVar9;
                iVar5 = FUN_180001510(plVar9);
                if (iVar5 == -1) {
                  uVar13 = 0xb5;
                }
                else {
                  if (iVar5 == 0) {
LAB_18004f5ad:
                    iVar5 = (int)*plVar14 + 1;
                    if (iVar5 != 0) {
                      *(int *)plVar14 = iVar5;
                    }
                    goto LAB_18004f085;
                  }
                  lVar6 = FUN_1800494f0(param_1);
                  if (lVar6 == 0) {
/* novels.agents.reviser_agent#90="apply_spot_fix_patches" */
                    FUN_180373340(param_1,&local_res20,DAT_1803aea70);
                    uVar13 = 0xb6;
                    lVar6 = local_res20;
                    plVar10 = local_res18;
                    goto LAB_18004f35e;
                  }
                  uVar13 = 0xb6;
                  *(undefined4 *)(plVar4 + 5) = 0xb6;
                  local_a8 = plVar14;
                  local_a0 = plVar9;
                  plVar9 = (longlong *)FUN_180377750(param_1,lVar6,&local_a8);
                  if (plVar9 != (longlong *)0x0) {
                    local_80 = plVar9;
                    plVar9 = (longlong *)FUN_18037f0c0();
                    if (plVar9 == (longlong *)0x0) {
                      uVar13 = 0xb7;
                    }
                    else {
                      iVar5 = FUN_180001510(plVar9);
                      if (iVar5 == -1) {
                        lVar6 = *(longlong *)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                          (**(code **)(plVar9[1] + 0x30))(plVar9);
                        }
                        uVar13 = 0xb7;
                        plVar10 = local_res18;
                        goto LAB_18004f35e;
                      }
                      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                      }
                      if (iVar5 == 0) goto LAB_18004f5ad;
                      plVar10 = (longlong *)FUN_18037f0c0();
                      if (plVar10 != (longlong *)0x0) goto LAB_18004f085;
                      uVar13 = 0xb8;
                    }
                  }
                }
              }
            }
          }
          else {
/* novels.agents.reviser_agent#1="auto" */
            iVar5 = FUN_1803836d0(plVar9,DAT_1803ae7a8);
            if (iVar5 == -1) {
              uVar13 = 0xb3;
            }
            else {
              if (iVar5 == 1) {
/* novels.agents.reviser_agent#88="patch-only" */
                iVar5 = FUN_1803836d0(plVar1,DAT_1803aea60);
                if (iVar5 == -1) {
                  uVar13 = 0xb3;
                  goto LAB_18004f34e;
                }
                if (iVar5 == 1) goto LAB_18004f1c9;
              }
/* novels.agents.reviser_agent#1="auto" */
              iVar5 = FUN_1803836d0(plVar9,DAT_1803ae7a8);
              if (iVar5 == -1) {
                uVar13 = 0xbb;
              }
              else if (iVar5 == 1) {
/* novels.agents.reviser_agent#93="rewrite-only" */
                iVar5 = FUN_1803836d0(plVar1,DAT_1803aea88);
                if (iVar5 == -1) {
                  uVar13 = 0xbb;
                }
                else {
                  if (iVar5 != 1) goto LAB_18004ed92;
                  iVar5 = FUN_180001510(plVar7);
                  if (iVar5 != -1) {
                    if (iVar5 == 1) {
                      plVar10 = plVar7;
                    }
                    iVar5 = (int)*plVar10 + 1;
                    if (iVar5 != 0) {
                      *(int *)plVar10 = iVar5;
                    }
                    goto LAB_18004f085;
                  }
                  uVar13 = 0xbc;
                }
              }
              else {
LAB_18004ed92:
                iVar5 = FUN_180001510(plVar7);
                if (iVar5 == -1) {
                  uVar13 = 0xbe;
                }
                else {
                  if (iVar5 != 0) {
                    iVar5 = (int)*plVar7 + 1;
                    plVar10 = plVar7;
                    if (iVar5 != 0) {
                      *(int *)plVar7 = iVar5;
                    }
LAB_18004f085:
                    FUN_180001d80(param_1);
LAB_18004f08d:
                    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                    }
                    if ((-1 < (int)*local_70) && (*local_70 = *local_70 + -1, *local_70 == 0)) {
                      (**(code **)(local_90[1] + 0x30))();
                    }
                    if (((local_88 != (longlong *)0x0) && (-1 < (int)*local_88)) &&
                       (*local_88 = *local_88 + -1, *local_88 == 0)) {
                      (**(code **)(local_88[1] + 0x30))();
                    }
                    if (((local_80 != (longlong *)0x0) && (-1 < (int)*local_80)) &&
                       (*local_80 = *local_80 + -1, *local_80 == 0)) {
                      (**(code **)(local_80[1] + 0x30))();
                    }
                    if (((local_res18 != (longlong *)0x0) && (-1 < (int)*local_res18)) &&
                       (*local_res18 = *local_res18 + -1, *local_res18 == 0)) {
                      (**(code **)(local_res18[1] + 0x30))();
                    }
                    if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
                      (**(code **)(plVar2[1] + 0x30))();
                    }
                    if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
                      (**(code **)(plVar14[1] + 0x30))(plVar14);
                    }
                    if ((-1 < (int)*local_78) && (*local_78 = *local_78 + -1, *local_78 == 0)) {
                      (**(code **)(local_78[1] + 0x30))();
                    }
                    if ((int)*plVar1 < 0) {
                      return plVar10;
                    }
                    *plVar1 = *plVar1 + -1;
                    if (*plVar1 != 0) {
                      return plVar10;
                    }
                    (**(code **)(plVar1[1] + 0x30))();
                    return plVar10;
                  }
                  lVar6 = FUN_180049d70(param_1);
                  if (lVar6 == 0) {
/* novels.agents.reviser_agent#89="parse_spot_fix_patches" */
                    FUN_180373340(param_1,&local_res20,DAT_1803aea68);
                    uVar13 = 0xc0;
                    lVar6 = local_res20;
                    plVar10 = local_res18;
                    goto LAB_18004f35e;
                  }
                  iVar5 = FUN_180001510(plVar8);
                  uVar13 = 0xc0;
                  if (iVar5 != -1) {
                    *(undefined4 *)(plVar4 + 5) = 0xc0;
                    plVar9 = plVar2;
                    if (iVar5 == 1) {
                      plVar9 = plVar8;
                    }
                    plVar9 = (longlong *)FUN_180376770(param_1,lVar6,plVar9);
                    if (plVar9 != (longlong *)0x0) {
                      local_88 = plVar9;
                      iVar5 = FUN_180001510(plVar9);
                      if (iVar5 == -1) {
                        uVar13 = 0xc1;
                      }
                      else if (iVar5 == 0) {
LAB_18004ef52:
                        plVar9 = (longlong *)FUN_18037f0c0();
                        uVar13 = 0xc5;
                        if (plVar9 != (longlong *)0x0) {
                          *(undefined4 *)(plVar4 + 5) = 0xc5;
                          plVar10 = (longlong *)FUN_180376210(param_1,plVar9);
                          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                            (**(code **)(plVar9[1] + 0x30))(plVar9);
                          }
                          plVar14 = local_a8;
                          if (plVar10 != (longlong *)0x0) {
                            plVar9 = (longlong *)FUN_18037f0c0();
                            uVar13 = 0xc6;
                            plVar14 = local_a8;
                            local_res18 = plVar10;
                            if (plVar9 != (longlong *)0x0) {
                              *(undefined4 *)(plVar4 + 5) = 0xc6;
                              plVar8 = (longlong *)FUN_1803770d0(param_1,plVar9,DAT_1803aea90);
                              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                (**(code **)(plVar9[1] + 0x30))(plVar9);
                              }
                              plVar14 = local_a8;
                              if (plVar8 != (longlong *)0x0) {
                                iVar5 = FUN_180001510(plVar8);
                                if (iVar5 != -1) {
                                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0))
                                  {
                                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                                  }
                                  if (iVar5 == 0) {
                                    FUN_180001d80(param_1);
                                    iVar5 = (int)*local_a8 + 1;
                                    plVar10 = local_a8;
                                    plVar14 = local_a8;
                                    if (iVar5 != 0) {
                                      *(int *)local_a8 = iVar5;
                                    }
                                    goto LAB_18004f08d;
                                  }
                                  iVar5 = (int)*plVar10 + 1;
                                  plVar14 = local_a8;
                                  if (iVar5 != 0) {
                                    *(int *)plVar10 = iVar5;
                                  }
                                  goto LAB_18004f085;
                                }
                                lVar6 = *(longlong *)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                                plVar14 = local_a8;
                                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                                  plVar14 = local_a8;
                                }
                                goto LAB_18004f35e;
                              }
                            }
                          }
                        }
                      }
                      else {
                        lVar6 = FUN_1800494f0(param_1);
                        if (lVar6 == 0) {
/* novels.agents.reviser_agent#90="apply_spot_fix_patches" */
                          FUN_180373340(param_1,&local_res20,DAT_1803aea70);
                          uVar13 = 0xc2;
                          lVar6 = local_res20;
                          plVar10 = local_res18;
                          goto LAB_18004f35e;
                        }
                        uVar13 = 0xc2;
                        *(undefined4 *)(plVar4 + 5) = 0xc2;
                        local_48 = plVar14;
                        local_40 = plVar9;
                        plVar9 = (longlong *)FUN_180377750(param_1,lVar6,&local_48);
                        if (plVar9 != (longlong *)0x0) {
                          local_80 = plVar9;
                          plVar9 = (longlong *)FUN_18037f0c0();
                          if (plVar9 == (longlong *)0x0) {
                            uVar13 = 0xc3;
                          }
                          else {
                            iVar5 = FUN_180001510(plVar9);
                            if (iVar5 == -1) {
                              lVar6 = *(longlong *)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                (**(code **)(plVar9[1] + 0x30))(plVar9);
                              }
                              uVar13 = 0xc3;
                              plVar10 = local_res18;
                              goto LAB_18004f35e;
                            }
                            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                            }
                            if (iVar5 == 0) goto LAB_18004ef52;
                            plVar10 = (longlong *)FUN_18037f0c0();
                            if (plVar10 != (longlong *)0x0) goto LAB_18004f085;
                            uVar13 = 0xc4;
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
LAB_18004f34e:
    lVar6 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    plVar10 = local_res18;
  }
LAB_18004f35e:
  piVar3 = *(int **)(lVar6 + 0x28);
  if (piVar3 == (int *)0x0) {
    pcVar11 = (code *)FUN_18038e2a0(plVar4,uVar13);
    plVar9 = *(longlong **)(lVar6 + 0x28);
    pcVar12 = pcVar11;
    if (pcVar11 == _Py_NoneStruct_exref) {
      pcVar12 = (code *)0x0;
    }
    else {
LAB_18004f3b8:
      pcVar11 = pcVar12;
      if ((pcVar12 != (code *)0x0) && (*(int *)pcVar12 + 1 != 0)) {
        *(int *)pcVar12 = *(int *)pcVar12 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar4) goto LAB_18004f401;
    pcVar11 = (code *)FUN_18038e2a0(plVar4,uVar13);
    *(int **)(pcVar11 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar9 = *(longlong **)(lVar6 + 0x28);
    pcVar12 = pcVar11;
    if (pcVar11 != _Py_NoneStruct_exref) goto LAB_18004f3b8;
    pcVar12 = (code *)0x0;
  }
  *(code **)(lVar6 + 0x28) = pcVar12;
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))();
  }
  if (((pcVar11 != (code *)0x0) && (-1 < *(int *)pcVar11)) &&
     (*(longlong *)pcVar11 = *(longlong *)pcVar11 + -1, *(longlong *)pcVar11 == 0)) {
    (**(code **)(*(longlong *)(pcVar11 + 8) + 0x30))(pcVar11);
  }
LAB_18004f401:
  plVar8 = local_80;
  plVar7 = local_88;
  FUN_18036d0a0(plVar4,"ooooooooo",plVar2,plVar14,local_78,plVar1,local_98,local_90,local_88,
                local_80,plVar10);
  plVar9 = DAT_1803ae750;
  if (plVar4 == DAT_1803ae750) {
    if ((-1 < (int)*DAT_1803ae750) && (*DAT_1803ae750 = *DAT_1803ae750 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))();
    }
    DAT_1803ae750 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((local_98 != (longlong *)0x0) && (-1 < (int)*local_98)) &&
     (*local_98 = *local_98 + -1, *local_98 == 0)) {
    (**(code **)(local_98[1] + 0x30))();
  }
  if (((local_90 != (longlong *)0x0) && (-1 < (int)*local_90)) &&
     (*local_90 = *local_90 + -1, *local_90 == 0)) {
    (**(code **)(local_90[1] + 0x30))();
  }
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))();
  }
  if ((-1 < (int)*plVar14) && (*plVar14 = *plVar14 + -1, *plVar14 == 0)) {
    (**(code **)(plVar14[1] + 0x30))(plVar14);
  }
  if ((-1 < (int)*local_78) && (*local_78 = *local_78 + -1, *local_78 == 0)) {
    (**(code **)(local_78[1] + 0x30))();
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar6;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (longlong *)0x0;
}
