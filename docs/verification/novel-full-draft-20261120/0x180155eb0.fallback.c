/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180155eb0 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

longlong FUN_180155eb0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  int *piVar1;
  undefined8 uVar2;
  char cVar3;
  int iVar4;
  longlong lVar5;
  longlong lVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  code *pcVar12;
  code *pcVar13;
  longlong *plVar14;
  undefined8 uVar15;
  longlong *plVar16;
  undefined1 auStack_158 [32];
  longlong *local_138;
  longlong *local_130;
  longlong *local_128;
  longlong *local_120;
  longlong *local_118;
  longlong *local_100;
  longlong *local_f8;
  longlong *local_f0;
  longlong *local_e8;
  longlong *local_e0;
  longlong *local_d8;
  longlong *local_d0;
  longlong *local_c8;
  longlong *local_c0;
  undefined8 local_b8;
  longlong *local_b0;
  undefined8 local_a8;
  longlong *local_a0;
  undefined8 local_98;
  longlong *local_90;
  undefined8 local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  ulonglong local_48;

  plVar16 = DAT_1803b6bb8;
  local_48 = DAT_1803a6040 ^ (ulonglong)auStack_158;
  plVar9 = (longlong *)*param_3;
  plVar10 = (longlong *)param_3[1];
  local_e0 = (longlong *)0x0;
  local_d8 = (longlong *)0x0;
  local_d0 = (longlong *)0x0;
  local_80 = (longlong *)0x0;
  local_78 = (longlong *)0x0;
  local_b0 = (longlong *)0x0;
  local_c0 = (longlong *)0x0;
  local_f8 = DAT_180391d20;
  local_e8 = DAT_1803b6bb8;
  local_f0 = plVar10;
  local_c8 = plVar9;
  if (DAT_1803b6bb8 == (longlong *)0x0) {
LAB_180155f51:
    DAT_1803b6bb8 = (longlong *)FUN_18036cd10(param_1,DAT_1803b6b08,DAT_1803ca758,0x38);
  }
  else if ((1 < *DAT_1803b6bb8) || (DAT_1803b6bb8[2] != 0)) {
    if ((DAT_1803b6bb8 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b6bb8 && (*DAT_1803b6bb8 = *DAT_1803b6bb8 + -1, *plVar16 == 0)))) {
      (**(code **)(plVar16[1] + 0x30))(plVar16);
    }
    goto LAB_180155f51;
  }
  plVar16 = DAT_1803b6bb8;
  local_e8 = DAT_1803b6bb8;
  FUN_180001d00(param_1,DAT_1803b6bb8);
  lVar5 = FUN_18014c240(param_1);
  if (lVar5 == 0) {
    FUN_180373340(param_1,&local_f8,DAT_1803b6d68);
    uVar15 = 0xd3;
    plVar9 = local_f8;
  }
  else {
    lVar6 = FUN_18014c020(param_1);
    if (lVar6 == 0) {
      FUN_180373340(param_1,&local_f8,DAT_1803b6da0);
      uVar15 = 0xd4;
      plVar9 = local_f8;
    }
    else {
      uVar15 = 0xd4;
      local_98 = DAT_1803b6e58;
      *(undefined4 *)(plVar16 + 5) = 0xd4;
      local_a0 = plVar9;
      plVar7 = (longlong *)FUN_180377750(param_1,lVar6,&local_a0);
      if (plVar7 == (longlong *)0x0) {
LAB_180156c8c:
        plVar9 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
        iVar4 = FUN_180001510(plVar7);
        if (iVar4 == -1) {
LAB_180156035:
          plVar9 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
        }
        else {
          if (iVar4 != 1) {
            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
              (**(code **)(plVar7[1] + 0x30))(plVar7);
            }
            lVar6 = FUN_18014c020(param_1);
            if (lVar6 == 0) {
              FUN_180373340(param_1,&local_f8,DAT_1803b6da0);
              uVar15 = 0xd5;
              plVar9 = local_f8;
              goto LAB_180156c98;
            }
            uVar15 = 0xd5;
            local_88 = DAT_1803b6db0;
            *(undefined4 *)(plVar16 + 5) = 0xd5;
            local_90 = plVar9;
            plVar7 = (longlong *)FUN_180377750(param_1,lVar6,&local_90);
            if (plVar7 == (longlong *)0x0) goto LAB_180156c8c;
            iVar4 = FUN_180001510(plVar7);
            if (iVar4 == -1) goto LAB_180156035;
            if (iVar4 != 1) {
              if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                (**(code **)(plVar7[1] + 0x30))(plVar7);
              }
              iVar4 = (int)*plVar10 + 1;
              plVar7 = plVar10;
              if (iVar4 != 0) {
                *(int *)plVar10 = iVar4;
              }
            }
          }
          uVar15 = 0xd3;
          *(undefined4 *)(plVar16 + 5) = 0xd3;
          plVar8 = (longlong *)FUN_180376770(param_1,lVar5,plVar7);
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          plVar10 = local_f0;
          if (plVar8 == (longlong *)0x0) goto LAB_180156c8c;
          iVar4 = FUN_180001510(plVar8);
          if (iVar4 == -1) {
            plVar9 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar10 = local_f0;
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
              plVar10 = local_f0;
            }
          }
          else {
            if (iVar4 != 1) {
              if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                (**(code **)(plVar8[1] + 0x30))(plVar8);
              }
              iVar4 = (int)*local_f0 + 1;
              if (iVar4 != 0) {
                *(int *)local_f0 = iVar4;
              }
              local_e0 = local_f0;
              plVar8 = local_e0;
            }
            local_e0 = plVar8;
            plVar10 = local_f0;
            lVar5 = FUN_18014c020(param_1);
            if (lVar5 != 0) {
              uVar15 = 0xd9;
              local_88 = DAT_1803b6e60;
              *(undefined4 *)(plVar16 + 5) = 0xd9;
              local_90 = plVar9;
              plVar7 = (longlong *)FUN_180377750(param_1,lVar5,&local_90);
              if (plVar7 != (longlong *)0x0) {
                iVar4 = FUN_180001510(plVar7);
                if (iVar4 == -1) goto LAB_180156035;
                if (iVar4 != 1) {
                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                  }
                  lVar5 = FUN_18014c020(param_1);
                  if (lVar5 == 0) {
                    FUN_180373340(param_1,&local_f8,DAT_1803b6da0);
                    uVar15 = 0xda;
                    plVar9 = local_f8;
                    goto LAB_180156c98;
                  }
                  uVar15 = 0xda;
                  local_98 = DAT_1803b6e68;
                  *(undefined4 *)(plVar16 + 5) = 0xda;
                  local_a0 = plVar9;
                  plVar7 = (longlong *)FUN_180377750(param_1,lVar5,&local_a0);
                  if (plVar7 == (longlong *)0x0) goto LAB_180156c8c;
                }
                plVar8 = (longlong *)FUN_18037f0c0();
                if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                }
                uVar15 = 0xd8;
                if (plVar8 != (longlong *)0x0) {
                  *(undefined4 *)(plVar16 + 5) = 0xd8;
                  plVar7 = (longlong *)FUN_180376210(param_1,plVar8);
                  local_100 = plVar7;
                  if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                    (**(code **)(plVar8[1] + 0x30))(plVar8);
                  }
                  if (plVar7 != (longlong *)0x0) {
                    local_d8 = plVar7;
                    lVar5 = FUN_18014c020(param_1);
                    if (lVar5 == 0) {
                      FUN_180373340(param_1,&local_f8,DAT_1803b6da0);
                      uVar15 = 0xdd;
                      plVar9 = local_f8;
                    }
                    else {
                      uVar15 = 0xdd;
                      local_88 = DAT_1803b6e70;
                      *(undefined4 *)(plVar16 + 5) = 0xdd;
                      local_90 = plVar9;
                      plVar7 = (longlong *)FUN_180377750(param_1,lVar5,&local_90);
                      if (plVar7 == (longlong *)0x0) goto LAB_180156c8c;
                      iVar4 = FUN_180001510(plVar7);
                      if (iVar4 == -1) goto LAB_180156035;
                      if (iVar4 != 1) {
                        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                        }
                        lVar5 = FUN_18014c020(param_1);
                        if (lVar5 == 0) {
                          FUN_180373340(param_1,&local_f8,DAT_1803b6da0);
                          uVar15 = 0xde;
                          plVar9 = local_f8;
                          goto LAB_180156c98;
                        }
                        uVar15 = 0xde;
                        local_98 = DAT_1803b6e78;
                        *(undefined4 *)(plVar16 + 5) = 0xde;
                        local_a0 = plVar9;
                        plVar7 = (longlong *)FUN_180377750(param_1,lVar5,&local_a0);
                        if (plVar7 == (longlong *)0x0) goto LAB_180156c8c;
                      }
                      local_d0 = plVar7;
                      plVar9 = (longlong *)FUN_18037f0c0();
                      if (plVar9 == (longlong *)0x0) {
                        plVar9 = *(longlong **)(param_1 + 0x60);
                        *(undefined8 *)(param_1 + 0x60) = 0;
                        uVar15 = 0xe0;
                        plVar10 = local_f0;
                      }
                      else {
                        *(undefined4 *)(plVar16 + 5) = 0xe2;
                        plVar10 = (longlong *)FUN_1803770d0(param_1,plVar9,DAT_1803b6c90);
                        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                          (**(code **)(plVar9[1] + 0x30))(plVar9);
                        }
                        if (plVar10 == (longlong *)0x0) {
                          plVar9 = *(longlong **)(param_1 + 0x60);
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          uVar15 = 0xe0;
                          plVar10 = local_f0;
                        }
                        else {
                          plVar7 = (longlong *)FUN_18000d280();
                          if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                            (**(code **)(plVar10[1] + 0x30))(plVar10);
                          }
                          if (plVar7 == (longlong *)0x0) {
                            plVar9 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            uVar15 = 0xe0;
                            plVar10 = local_f0;
                          }
                          else {
                            plVar8 = (longlong *)FUN_180372160(param_1,0);
                            do {
                              plVar11 = local_b0;
                              plVar10 = local_c0;
                              pcVar12 = *(code **)(plVar7[1] + 0xe0);
                              if (pcVar12 == (code *)0x0) {
                                PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,
                                             "\'%s\' object is not an iterator",
                                             *(undefined8 *)(plVar7[1] + 0x18));
LAB_180156868:
                                plVar9 = *(longlong **)(param_1 + 0x60);
                                if (plVar9 != (longlong *)0x0) {
                                  plVar14 = (longlong *)plVar9[1];
                                  if ((*(uint *)((longlong *)plVar9[1] + 0x15) & 0x40000000) == 0) {
                                    plVar14 = plVar9;
                                  }
                                  if ((plVar14 != *(longlong **)PyExc_StopIteration_exref) &&
                                     (((-1 < *(int *)(plVar14[1] + 0xa8) ||
                                       ((*(uint *)(plVar14 + 0x15) & 0x40000000) == 0)) ||
                                      (cVar3 = FUN_180382640(), cVar3 != '\x01'))))
                                  goto LAB_1801568c0;
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0))
                                  {
                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                  }
                                }
                                iVar4 = (int)*plVar8 + 1;
                                if (iVar4 != 0) {
                                  *(int *)plVar8 = iVar4;
                                }
                                if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                                }
                                if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                  (**(code **)(plVar8[1] + 0x30))(plVar8);
                                }
                                if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
                                   (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                                }
                                if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
                                   (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                  (**(code **)(plVar11[1] + 0x30))(plVar11);
                                }
                                local_80 = plVar8;
                                lVar5 = FUN_18014c020(param_1);
                                plVar9 = local_c8;
                                uVar2 = DAT_1803b6e90;
                                if (lVar5 == 0) {
                                  FUN_180373340(param_1,&local_f8,DAT_1803b6da0);
                                  uVar15 = 0xe6;
                                  plVar9 = local_f8;
                                  plVar10 = local_f0;
                                  goto LAB_180156c98;
                                }
                                uVar15 = 0xe6;
                                *(undefined4 *)(plVar16 + 5) = 0xe6;
                                local_b0 = local_c8;
                                local_a8 = uVar2;
                                plVar7 = (longlong *)FUN_180377750(param_1,lVar5,&local_b0);
                                plVar10 = local_f0;
                                if (plVar7 == (longlong *)0x0) goto LAB_180156c8c;
                                iVar4 = FUN_180001510(plVar7);
                                if (iVar4 == -1) {
                                  plVar9 = *(longlong **)(param_1 + 0x60);
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  plVar10 = local_f0;
                                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0))
                                  {
                                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                                    plVar10 = local_f0;
                                  }
                                  goto LAB_180156c98;
                                }
                                if (iVar4 != 1) {
                                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0))
                                  {
                                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                                  }
                                  lVar5 = FUN_18014c020(param_1);
                                  if (lVar5 == 0) {
                                    FUN_180373340(param_1,&local_f8,DAT_1803b6da0);
                                    uVar15 = 0xe7;
                                    plVar9 = local_f8;
                                    plVar10 = local_f0;
                                    goto LAB_180156c98;
                                  }
                                  uVar15 = 0xe7;
                                  local_b8 = DAT_1803b6e98;
                                  *(undefined4 *)(plVar16 + 5) = 0xe7;
                                  local_c0 = plVar9;
                                  plVar7 = (longlong *)FUN_180377750(param_1,lVar5,&local_c0);
                                  plVar10 = local_f0;
                                  if (plVar7 == (longlong *)0x0) goto LAB_180156c8c;
                                }
                                plVar9 = (longlong *)FUN_18037f0c0();
                                if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                                }
                                uVar15 = 0xe5;
                                plVar10 = local_f0;
                                if (plVar9 == (longlong *)0x0) goto LAB_180156c8c;
                                *(undefined4 *)(plVar16 + 5) = 0xe5;
                                plVar7 = (longlong *)FUN_180376210(param_1,plVar9);
                                if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                  (**(code **)(plVar9[1] + 0x30))(plVar9);
                                }
                                plVar10 = local_f0;
                                if (plVar7 == (longlong *)0x0) goto LAB_180156c8c;
                                local_78 = plVar7;
                                lVar5 = FUN_18014bd60(param_1);
                                if (lVar5 != 0) {
                                  plVar9 = (longlong *)FUN_18037f0c0();
                                  uVar15 = 0xee;
                                  plVar16 = local_e8;
                                  plVar10 = local_f0;
                                  if (plVar9 != (longlong *)0x0) {
                                    *(undefined4 *)(local_e8 + 5) = 0xee;
                                    plVar11 = (longlong *)FUN_180376210(param_1,plVar9);
                                    if ((-1 < (int)*plVar9) &&
                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                      (**(code **)(plVar9[1] + 0x30))(plVar9);
                                    }
                                    plVar16 = local_e8;
                                    plVar10 = local_f0;
                                    if (plVar11 != (longlong *)0x0) {
                                      uVar15 = 0xe9;
                                      *(undefined4 *)(local_e8 + 5) = 0xe9;
                                      local_70 = local_e0;
                                      local_68 = local_d8;
                                      local_60 = plVar8;
                                      local_58 = plVar7;
                                      local_50 = plVar11;
                                      lVar5 = FUN_18037ccb0(param_1,lVar5,&local_70,DAT_1803b6ea8);
                                      if ((-1 < (int)*plVar11) &&
                                         (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                        (**(code **)(plVar11[1] + 0x30))(plVar11);
                                      }
                                      plVar16 = local_e8;
                                      plVar10 = local_f0;
                                      if (lVar5 != 0) {
                                        FUN_180001d80(param_1);
                                        if ((-1 < (int)*local_e0) &&
                                           (*local_e0 = *local_e0 + -1, *local_e0 == 0)) {
                                          (**(code **)(local_e0[1] + 0x30))();
                                        }
                                        if ((-1 < (int)*local_100) &&
                                           (*local_100 = *local_100 + -1, *local_100 == 0)) {
                                          (**(code **)(local_d8[1] + 0x30))();
                                        }
                                        if ((-1 < (int)*local_d0) &&
                                           (*local_d0 = *local_d0 + -1, *local_d0 == 0)) {
                                          (**(code **)(local_d0[1] + 0x30))();
                                        }
                                        if ((-1 < (int)*plVar8) &&
                                           (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                                          (**(code **)(plVar8[1] + 0x30))(plVar8);
                                        }
                                        if ((-1 < (int)*plVar7) &&
                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                                        }
                                        if ((-1 < (int)*local_c8) &&
                                           (*local_c8 = *local_c8 + -1, *local_c8 == 0)) {
                                          (**(code **)(local_c8[1] + 0x30))();
                                        }
                                        if ((int)*local_f0 < 0) {
                                          return lVar5;
                                        }
                                        *local_f0 = *local_f0 + -1;
                                        if (*local_f0 != 0) {
                                          return lVar5;
                                        }
                                        (**(code **)(local_f0[1] + 0x30))();
                                        return lVar5;
                                      }
                                    }
                                  }
                                  goto LAB_180156c8c;
                                }
                                FUN_180373340(param_1,&local_f8,DAT_1803b6ea0);
                                uVar15 = 0xe9;
                                plVar9 = local_f8;
                                plVar16 = local_e8;
                                plVar10 = local_f0;
                                goto LAB_180156c98;
                              }
                              plVar9 = (longlong *)(*pcVar12)(plVar7);
                              if (plVar9 == (longlong *)0x0) goto LAB_180156868;
                              local_c0 = plVar9;
                              if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
                                 (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                              }
                              iVar4 = (int)*plVar9 + 1;
                              if (iVar4 != 0) {
                                *(int *)plVar9 = iVar4;
                              }
                              local_b0 = plVar9;
                              if (((plVar11 != (longlong *)0x0) && (-1 < (int)*plVar11)) &&
                                 (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                (**(code **)(plVar11[1] + 0x30))();
                              }
                              lVar5 = FUN_18014c500(param_1);
                              if (lVar5 == 0) {
LAB_1801567f9:
                                FUN_180373340(param_1,&local_f8,DAT_1803b6c30);
                                plVar9 = local_f8;
                                goto LAB_1801568c8;
                              }
                              local_70 = DAT_1803b6e88;
                              local_68 = DAT_1803b6c50;
                              *(undefined4 *)(plVar16 + 5) = 0xe3;
                              local_60 = plVar9;
                              plVar10 = (longlong *)
                                        FUN_18037e980(param_1,lVar5,DAT_1803b6d18,&local_70);
                              if (plVar10 == (longlong *)0x0) goto LAB_1801567f0;
                              plVar11 = (longlong *)FUN_18037f0c0();
                              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                              {
                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                              }
                              if (plVar11 == (longlong *)0x0) goto LAB_1801567f0;
                              *(undefined4 *)(plVar16 + 5) = 0xe3;
                              plVar10 = (longlong *)FUN_180376210(param_1,plVar11);
                              if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0))
                              {
                                (**(code **)(plVar11[1] + 0x30))(plVar11);
                              }
                              if (plVar10 == (longlong *)0x0) goto LAB_1801567f0;
                              iVar4 = FUN_180001510(plVar10);
                              if (iVar4 == -1) {
                                plVar9 = *(longlong **)(param_1 + 0x60);
                                *(undefined8 *)(param_1 + 0x60) = 0;
                                if ((-1 < (int)*plVar10) &&
                                   (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                                }
                                goto LAB_1801568c8;
                              }
                              if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0))
                              {
                                (**(code **)(plVar10[1] + 0x30))(plVar10);
                              }
                              if (iVar4 != 0) {
                                lVar5 = FUN_18014c500(param_1);
                                if (lVar5 == 0) goto LAB_1801567f9;
                                local_70 = DAT_1803b6e88;
                                local_68 = DAT_1803b6c50;
                                *(undefined4 *)(plVar16 + 5) = 0xe1;
                                local_60 = plVar9;
                                plVar9 = (longlong *)
                                         FUN_18037e980(param_1,lVar5,DAT_1803b6d18,&local_70);
                                if (plVar9 == (longlong *)0x0) goto LAB_1801567f0;
                                plVar10 = (longlong *)FUN_18037f0c0();
                                if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                  (**(code **)(plVar9[1] + 0x30))(plVar9);
                                }
                                if (plVar10 == (longlong *)0x0) goto LAB_1801567f0;
                                *(undefined4 *)(plVar16 + 5) = 0xe1;
                                lVar5 = FUN_180376210(param_1,plVar10);
                                if ((-1 < (int)*plVar10) &&
                                   (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                                }
                                if (lVar5 == 0) goto LAB_1801567f0;
                                lVar6 = plVar8[2];
                                cVar3 = FUN_1803722b0(plVar8,lVar6 + 1);
                                if (cVar3 == '\0') goto LAB_1801567f0;
                                *(longlong *)(plVar8[3] + lVar6 * 8) = lVar5;
                              }
                              lVar5 = *(longlong *)(param_1 + 0x10);
                              if ((((*(int *)(*(longlong *)(lVar5 + 0x60) + 0x4d4) != 0) ||
                                   (*(int *)(lVar5 + 0x1a0) != 0)) &&
                                  (iVar4 = Py_MakePendingCalls(), iVar4 < 0)) &&
                                 (*(longlong *)(param_1 + 0x60) != 0)) goto LAB_1801567f0;
                              if (*(int *)(lVar5 + 0x174) != 0) {
                                PyEval_SaveThread();
                                PyEval_AcquireThread(param_1);
                              }
                            } while (*(longlong *)(param_1 + 0x80) == 0);
                            plVar9 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x80) = 0;
                            uVar15 = FUN_180373490(param_1);
                            *(undefined8 *)(param_1 + 0x60) = uVar15;
                            if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
                               (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                              (**(code **)(plVar9[1] + 0x30))(plVar9);
                            }
LAB_1801567f0:
                            plVar9 = *(longlong **)(param_1 + 0x60);
LAB_1801568c0:
                            *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1801568c8:
                            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                            }
                            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
                              (**(code **)(plVar8[1] + 0x30))(plVar8);
                            }
                            if (((local_c0 != (longlong *)0x0) && (-1 < (int)*local_c0)) &&
                               (*local_c0 = *local_c0 + -1, *local_c0 == 0)) {
                              (**(code **)(local_c0[1] + 0x30))();
                            }
                            if (((local_b0 != (longlong *)0x0) && (-1 < (int)*local_b0)) &&
                               (*local_b0 = *local_b0 + -1, *local_b0 == 0)) {
                              (**(code **)(local_b0[1] + 0x30))();
                            }
                            uVar15 = 0xe0;
                            plVar10 = local_f0;
                          }
                        }
                      }
                    }
                    goto LAB_180156c98;
                  }
                }
              }
              goto LAB_180156c8c;
            }
            FUN_180373340(param_1,&local_f8,DAT_1803b6da0);
            uVar15 = 0xd9;
            plVar9 = local_f8;
          }
        }
      }
    }
  }
LAB_180156c98:
  piVar1 = (int *)plVar9[5];
  if (piVar1 == (int *)0x0) {
    pcVar12 = (code *)FUN_18038e2a0(plVar16,uVar15);
    plVar7 = (longlong *)plVar9[5];
    pcVar13 = pcVar12;
    if (pcVar12 == _Py_NoneStruct_exref) {
      pcVar13 = (code *)0x0;
    }
    else {
LAB_180156cf2:
      pcVar12 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar16) goto LAB_180156d3b;
    pcVar12 = (code *)FUN_18038e2a0(plVar16,uVar15);
    *(int **)(pcVar12 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
    plVar7 = (longlong *)plVar9[5];
    pcVar13 = pcVar12;
    if (pcVar12 != _Py_NoneStruct_exref) goto LAB_180156cf2;
    pcVar13 = (code *)0x0;
  }
  plVar9[5] = (longlong)pcVar13;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
  }
LAB_180156d3b:
  plVar11 = local_c8;
  plVar8 = local_e0;
  local_118 = local_78;
  local_120 = local_80;
  local_128 = local_d0;
  local_130 = local_d8;
  local_138 = local_e0;
  FUN_18036d0a0(plVar16,"ooooooo",local_c8,plVar10);
  plVar7 = DAT_1803b6bb8;
  if (plVar16 == DAT_1803b6bb8) {
    if ((-1 < (int)*DAT_1803b6bb8) && (*DAT_1803b6bb8 = *DAT_1803b6bb8 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))();
    }
    DAT_1803b6bb8 = (longlong *)0x0;
  }
  lVar5 = **(longlong **)(param_1 + 0x38);
  plVar16 = *(longlong **)(lVar5 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar5 + 8);
  *(undefined4 *)(plVar16 + 8) = 0xffffffff;
  plVar7 = (longlong *)plVar16[2];
  if (((plVar7 != (longlong *)0x0) && (plVar16[2] = 0, -1 < (int)*plVar7)) &&
     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))(plVar16);
  }
  plVar16[10] = 0;
  if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)
     ) {
    (**(code **)(plVar8[1] + 0x30))(plVar8);
  }
  if (((local_d8 != (longlong *)0x0) && (-1 < (int)*local_d8)) &&
     (*local_d8 = *local_d8 + -1, *local_d8 == 0)) {
    (**(code **)(local_d8[1] + 0x30))();
  }
  if (((local_d0 != (longlong *)0x0) && (-1 < (int)*local_d0)) &&
     (*local_d0 = *local_d0 + -1, *local_d0 == 0)) {
    (**(code **)(local_d0[1] + 0x30))();
  }
  if (((local_80 != (longlong *)0x0) && (-1 < (int)*local_80)) &&
     (*local_80 = *local_80 + -1, *local_80 == 0)) {
    (**(code **)(local_80[1] + 0x30))();
  }
  if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
     (*local_78 = *local_78 + -1, *local_78 == 0)) {
    (**(code **)(local_78[1] + 0x30))();
  }
  if ((-1 < (int)*plVar11) && (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
    (**(code **)(plVar11[1] + 0x30))(plVar11);
  }
  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  plVar10 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar9;
  if (((plVar10 != (longlong *)0x0) && (-1 < (int)*plVar10)) &&
     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))();
  }
  return 0;
}
