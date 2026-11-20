code * FUN_1800d76c0(longlong param_1,undefined8 param_2,undefined8 *param_3)
{
  longlong *plVar1;
  int *piVar2;
  longlong *plVar3;
  undefined8 uVar4;
  int iVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong lVar8;
  longlong *plVar9;
  longlong *plVar10;
  longlong *plVar11;
  code *pcVar12;
  code *pcVar13;
  longlong lVar14;
  undefined8 uVar15;
  longlong lVar16;
  longlong *local_res20;
  longlong local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  undefined8 local_50;
  longlong *local_48;
  longlong *local_40;
  plVar3 = DAT_1803b3610;
  local_88 = (longlong *)*param_3;
  lVar8 = 0;
  plVar1 = (longlong *)param_3[1];
  plVar10 = (longlong *)param_3[2];
  local_70 = (longlong *)param_3[3];
  local_40 = (longlong *)param_3[4];
  local_90 = 0;
  local_80 = (longlong *)0x0;
  local_78 = (longlong *)0x0;
  local_48 = (longlong *)0x0;
  local_res20 = plVar10;
  if (DAT_1803b3610 == (longlong *)0x0) {
LAB_1800d7752:
    DAT_1803b3610 = (longlong *)FUN_18036cd10(param_1,DAT_1803b29a0,DAT_1803ca7a0,0x40);
  }
  else if ((1 < *DAT_1803b3610) || (DAT_1803b3610[2] != 0)) {
    if ((DAT_1803b3610 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b3610 && (*DAT_1803b3610 = *DAT_1803b3610 + -1, *plVar3 == 0)))) {
      (**(code **)(plVar3[1] + 0x30))(plVar3);
    }
    goto LAB_1800d7752;
  }
  plVar3 = DAT_1803b3610;
  FUN_180001d00(param_1,DAT_1803b3610);
  lVar14 = *(longlong *)(DAT_1803ca798 + 0x20);
  if (*(char *)(lVar14 + 10) == '\0') {
    plVar6 = (longlong *)FUN_1800015a0(DAT_1803ca798,DAT_1803b3278);
    if (plVar6 != (longlong *)0x0) {
      lVar8 = *plVar6;
    }
LAB_1800d785c:
    if (lVar8 == 0) goto LAB_1800d7861;
LAB_1800d78a2:
    plVar6 = (longlong *)FUN_18037f0c0();
    uVar15 = 0x207;
    if (plVar6 == (longlong *)0x0) {
      lVar8 = *(longlong *)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      *(undefined4 *)(plVar3 + 5) = 0x207;
      plVar7 = (longlong *)FUN_180376770(param_1,lVar8,plVar6);
      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
        (**(code **)(plVar6[1] + 0x30))(plVar6);
      }
      if (plVar7 == (longlong *)0x0) {
        lVar8 = *(longlong *)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
        iVar5 = FUN_180001510(plVar7);
        if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
          (**(code **)(plVar7[1] + 0x30))(plVar7);
        }
        pcVar12 = _Py_NoneStruct_exref;
        if (iVar5 == -1) {
          lVar8 = *(longlong *)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          if (iVar5 == 0) {
            if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
              *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
            }
            FUN_180001d80(param_1);
            goto LAB_1800d8666;
          }
          lVar8 = *(longlong *)(DAT_1803ca798 + 0x20);
          if (*(char *)(lVar8 + 10) == '\0') {
            plVar6 = (longlong *)FUN_1800015a0();
            if (plVar6 == (longlong *)0x0) {
              lVar14 = 0;
            }
            else {
              lVar14 = *plVar6;
            }
LAB_1800d7a3a:
            if (lVar14 == 0) goto LAB_1800d7a3f;
          }
          else {
            iVar5 = *(int *)(lVar8 + 0xc);
            if (iVar5 == 0) {
              iVar5 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
              *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar5 + 1;
              *(int *)(lVar8 + 0xc) = iVar5;
            }
            if (iVar5 != DAT_1803a9794) {
              DAT_1803a9794 = iVar5;
              DAT_1803b3608 =
                   FUN_1803707d0(lVar8,DAT_1803b3280,*(undefined8 *)(DAT_1803b3280 + 0x18));
            }
            if (-1 < DAT_1803b3608) {
              lVar16 = (1L << (*(byte *)(lVar8 + 9) & 0x3f)) + lVar8;
              lVar14 = *(longlong *)(lVar16 + 0x28 + DAT_1803b3608 * 0x10);
              if (lVar14 != 0) goto LAB_1800d7a80;
              DAT_1803b3608 =
                   FUN_1803707d0(lVar8,DAT_1803b3280,*(undefined8 *)(DAT_1803b3280 + 0x18));
              if (-1 < DAT_1803b3608) {
                lVar14 = *(longlong *)(lVar16 + 0x28 + DAT_1803b3608 * 0x10);
                goto LAB_1800d7a3a;
              }
            }
LAB_1800d7a3f:
            plVar6 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b3280);
            if ((plVar6 == (longlong *)0x0) || (lVar14 = *plVar6, lVar14 == 0)) {
              FUN_180373340(param_1,&local_90,DAT_1803b3280);
              uVar15 = 0x209;
              lVar8 = local_90;
              goto LAB_1800d8366;
            }
          }
LAB_1800d7a80:
          uVar15 = 0x209;
          local_60 = DAT_1803b3288;
          *(undefined4 *)(plVar3 + 5) = 0x209;
          local_68 = plVar1;
          plVar6 = (longlong *)FUN_180377750(param_1,lVar14,&local_68);
          if (plVar6 == (longlong *)0x0) {
            lVar8 = *(longlong *)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
          }
          else {
            iVar5 = FUN_180001510(plVar6);
            if (iVar5 == -1) {
              lVar8 = *(longlong *)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                (**(code **)(plVar6[1] + 0x30))(plVar6);
              }
            }
            else {
              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                (**(code **)(plVar6[1] + 0x30))(plVar6);
              }
              pcVar12 = _Py_NoneStruct_exref;
              if (iVar5 != 0) {
                if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
                  *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
                }
                FUN_180001d80(param_1);
                goto LAB_1800d8666;
              }
              lVar8 = FUN_1800c7020(param_1);
              if (lVar8 == 0) {
                FUN_180373340(param_1,&local_90,DAT_1803b2b20);
                uVar15 = 0x20b;
                lVar8 = local_90;
              }
              else {
                plVar6 = (longlong *)FUN_18037f0c0();
                if (plVar6 == (longlong *)0x0) {
                  lVar8 = *(longlong *)(param_1 + 0x60);
                  uVar15 = 0x20b;
                  *(undefined8 *)(param_1 + 0x60) = 0;
                }
                else {
                  plVar7 = (longlong *)FUN_180382540(plVar1,DAT_1803b2c90);
                  if (plVar7 == (longlong *)0x0) {
                    lVar8 = *(longlong *)(param_1 + 0x60);
                    *(undefined8 *)(param_1 + 0x60) = 0;
                    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                    }
                    uVar15 = 0x20b;
                  }
                  else {
                    plVar9 = (longlong *)FUN_180382540(plVar7,DAT_1803b2f70);
                    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                    }
                    if (plVar9 == (longlong *)0x0) {
                      lVar8 = *(longlong *)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                        (**(code **)(plVar6[1] + 0x30))(plVar6);
                      }
                      uVar15 = 0x20b;
                    }
                    else {
                      uVar15 = 0x20b;
                      local_60 = DAT_1803b2b10;
                      *(undefined4 *)(plVar3 + 5) = 0x20b;
                      local_68 = plVar9;
                      plVar10 = (longlong *)FUN_180377750(param_1,plVar6,&local_68);
                      if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                        (**(code **)(plVar6[1] + 0x30))(plVar6);
                      }
                      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                      }
                      if (plVar10 != (longlong *)0x0) {
                        plVar6 = (longlong *)FUN_18037f0c0();
                        if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                          (**(code **)(plVar10[1] + 0x30))(plVar10);
                        }
                        if (plVar6 != (longlong *)0x0) {
                          *(undefined4 *)(plVar3 + 5) = 0x20b;
                          plVar7 = (longlong *)FUN_180376210(param_1,plVar6);
                          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                            (**(code **)(plVar6[1] + 0x30))(plVar6);
                          }
                          if (plVar7 != (longlong *)0x0) {
                            local_80 = plVar7;
                            iVar5 = FUN_180001510(plVar7);
                            pcVar12 = _Py_NoneStruct_exref;
                            if (iVar5 != -1) {
                              if (iVar5 == 0) {
                                plVar10 = local_res20;
                                if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
                                  *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
                                }
LAB_1800d8606:
                                FUN_180001d80(param_1);
                                if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                  (**(code **)(plVar7[1] + 0x30))(plVar7);
                                }
                                if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
                                   (*local_78 = *local_78 + -1, *local_78 == 0)) {
                                  (**(code **)(local_78[1] + 0x30))();
                                }
                                if (((local_48 != (longlong *)0x0) && (-1 < (int)*local_48)) &&
                                   (*local_48 = *local_48 + -1, *local_48 == 0)) {
                                  (**(code **)(local_48[1] + 0x30))();
                                }
LAB_1800d8666:
                                if ((-1 < (int)*local_88) &&
                                   (*local_88 = *local_88 + -1, *local_88 == 0)) {
                                  (**(code **)(local_88[1] + 0x30))();
                                }
                                if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
                                  (**(code **)(plVar1[1] + 0x30))(plVar1);
                                }
                                if ((-1 < (int)*plVar10) &&
                                   (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                                }
                                if ((-1 < (int)*local_70) &&
                                   (*local_70 = *local_70 + -1, *local_70 == 0)) {
                                  (**(code **)(local_70[1] + 0x30))();
                                }
                                if ((int)*local_40 < 0) {
                                  return pcVar12;
                                }
                                *local_40 = *local_40 + -1;
                                if (*local_40 != 0) {
                                  return pcVar12;
                                }
                                (**(code **)(local_40[1] + 0x30))(local_40);
                                return pcVar12;
                              }
                              plVar10 = (longlong *)FUN_180370dc0(param_1,local_70);
                              if (plVar10 == (longlong *)0x0) {
                                uVar15 = 0x20e;
                                goto LAB_1800d7c88;
                              }
                              lVar8 = *(longlong *)(DAT_1803ca798 + 0x20);
                              local_78 = plVar10;
                              if (*(char *)(lVar8 + 10) == '\0') {
                                plVar6 = (longlong *)FUN_1800015a0();
                                if (plVar6 == (longlong *)0x0) {
                                  lVar14 = 0;
                                }
                                else {
                                  lVar14 = *plVar6;
                                }
LAB_1800d7e2d:
                                if (lVar14 == 0) goto LAB_1800d7e32;
                              }
                              else {
                                iVar5 = *(int *)(lVar8 + 0xc);
                                if (iVar5 == 0) {
                                  iVar5 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                                  *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar5 + 1;
                                  *(int *)(lVar8 + 0xc) = iVar5;
                                }
                                if (iVar5 != DAT_1803a9768) {
                                  DAT_1803a9768 = iVar5;
                                  DAT_1803b2a98 =
                                       FUN_1803707d0(lVar8,DAT_1803b3290,
                                                     *(undefined8 *)(DAT_1803b3290 + 0x18));
                                }
                                if (-1 < DAT_1803b2a98) {
                                  lVar16 = (1L << (*(byte *)(lVar8 + 9) & 0x3f)) + lVar8;
                                  lVar14 = *(longlong *)(lVar16 + 0x28 + DAT_1803b2a98 * 0x10);
                                  if (lVar14 != 0) goto LAB_1800d7e77;
                                  DAT_1803b2a98 =
                                       FUN_1803707d0(lVar8,DAT_1803b3290,
                                                     *(undefined8 *)(DAT_1803b3290 + 0x18));
                                  if (-1 < DAT_1803b2a98) {
                                    lVar14 = *(longlong *)(lVar16 + 0x28 + DAT_1803b2a98 * 0x10);
                                    goto LAB_1800d7e2d;
                                  }
                                }
LAB_1800d7e32:
                                plVar6 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b3290);
                                if ((plVar6 == (longlong *)0x0) || (lVar14 = *plVar6, lVar14 == 0))
                                {
                                  FUN_180373340(param_1,&local_90,DAT_1803b3290);
                                  uVar15 = 0x20f;
                                  lVar8 = local_90;
                                  plVar10 = local_res20;
                                  goto LAB_1800d8366;
                                }
                              }
LAB_1800d7e77:
                              uVar15 = 0x20f;
                              *(undefined4 *)(plVar3 + 5) = 0x20f;
                              local_68 = plVar1;
                              local_60 = plVar10;
                              plVar10 = (longlong *)FUN_180377750(param_1,lVar14,&local_68);
                              if (plVar10 != (longlong *)0x0) {
                                if ((-1 < (int)*plVar10) &&
                                   (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                                }
                                plVar10 = (longlong *)FUN_18037f0c0();
                                uVar15 = 0x210;
                                if (plVar10 != (longlong *)0x0) {
                                  *(undefined4 *)(plVar3 + 5) = 0x210;
                                  plVar6 = (longlong *)FUN_1803770d0(param_1,plVar10,DAT_1803b2b40);
                                  if ((-1 < (int)*plVar10) &&
                                     (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                                    (**(code **)(plVar10[1] + 0x30))(plVar10);
                                  }
                                  if (plVar6 != (longlong *)0x0) {
                                    iVar5 = FUN_180001510(plVar6);
                                    if (iVar5 == -1) {
                                      lVar8 = *(longlong *)(param_1 + 0x60);
                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                      plVar10 = local_res20;
                                      if ((-1 < (int)*plVar6) &&
                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                        (**(code **)(plVar6[1] + 0x30))(plVar6);
                                        plVar10 = local_res20;
                                      }
                                    }
                                    else {
                                      if (iVar5 != 1) {
                                        if ((-1 < (int)*plVar6) &&
                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                          (**(code **)(plVar6[1] + 0x30))(plVar6);
                                        }
                                        iVar5 = (int)*local_res20 + 1;
                                        plVar6 = local_res20;
                                        if (iVar5 != 0) {
                                          *(int *)local_res20 = iVar5;
                                        }
                                      }
                                      plVar10 = local_res20;
                                      plVar7 = (longlong *)FUN_1803749a0(plVar6);
                                      if ((-1 < (int)*plVar6) &&
                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                        (**(code **)(plVar6[1] + 0x30))(plVar6);
                                      }
                                      if (plVar7 == (longlong *)0x0) {
                                        lVar8 = *(longlong *)(param_1 + 0x60);
                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                      }
                                      else {
                                        lVar8 = *(longlong *)(DAT_1803ca798 + 0x20);
                                        local_48 = plVar7;
                                        if (*(char *)(lVar8 + 10) == '\0') {
                                          plVar6 = (longlong *)FUN_1800015a0();
                                          if (plVar6 == (longlong *)0x0) {
                                            lVar14 = 0;
                                          }
                                          else {
                                            lVar14 = *plVar6;
                                          }
LAB_1800d809a:
                                          if (lVar14 == 0) goto LAB_1800d809f;
                                        }
                                        else {
                                          iVar5 = *(int *)(lVar8 + 0xc);
                                          if (iVar5 == 0) {
                                            iVar5 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118
                                                            );
                                            *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) =
                                                 iVar5 + 1;
                                            *(int *)(lVar8 + 0xc) = iVar5;
                                          }
                                          if (iVar5 != DAT_1803a978c) {
                                            DAT_1803a978c = iVar5;
                                            DAT_1803b2968 =
                                                 FUN_1803707d0(lVar8,DAT_1803b3298,
                                                               *(undefined8 *)(DAT_1803b3298 + 0x18)
                                                              );
                                          }
                                          if (-1 < DAT_1803b2968) {
                                            lVar16 = (1L << (*(byte *)(lVar8 + 9) & 0x3f)) + lVar8;
                                            lVar14 = *(longlong *)
                                                      (lVar16 + 0x28 + DAT_1803b2968 * 0x10);
                                            if (lVar14 != 0) goto LAB_1800d80e0;
                                            DAT_1803b2968 =
                                                 FUN_1803707d0(lVar8,DAT_1803b3298,
                                                               *(undefined8 *)(DAT_1803b3298 + 0x18)
                                                              );
                                            if (-1 < DAT_1803b2968) {
                                              lVar14 = *(longlong *)
                                                        (lVar16 + 0x28 + DAT_1803b2968 * 0x10);
                                              goto LAB_1800d809a;
                                            }
                                          }
LAB_1800d809f:
                                          plVar6 = (longlong *)
                                                   FUN_1800015a0(DAT_1803ab778,DAT_1803b3298);
                                          if ((plVar6 == (longlong *)0x0) ||
                                             (lVar14 = *plVar6, lVar14 == 0)) {
                                            FUN_180373340(param_1,&local_90,DAT_1803b3298);
                                            uVar15 = 0x211;
                                            lVar8 = local_90;
                                            goto LAB_1800d8366;
                                          }
                                        }
LAB_1800d80e0:
                                        local_60 = DAT_1803b3288;
                                        uVar15 = 0x211;
                                        *(undefined4 *)(plVar3 + 5) = 0x211;
                                        local_68 = plVar10;
                                        local_res20 = plVar1;
                                        local_58 = plVar7;
                                        plVar6 = (longlong *)
                                                 FUN_18037d030(param_1,lVar14,&local_res20,&local_68
                                                               ,DAT_1803b32a0);
                                        if (plVar6 == (longlong *)0x0) {
                                          lVar8 = *(longlong *)(param_1 + 0x60);
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                        }
                                        else {
                                          if ((-1 < (int)*plVar6) &&
                                             (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                            (**(code **)(plVar6[1] + 0x30))(plVar6);
                                          }
                                          uVar15 = 0x212;
                                          *(undefined4 *)(plVar3 + 5) = 0x212;
                                          plVar6 = (longlong *)
                                                   FUN_18037e980(param_1,local_88,DAT_1803b2ee0,
                                                                 DAT_1803b32a8 + 0x18);
                                          if (plVar6 == (longlong *)0x0) {
                                            lVar8 = *(longlong *)(param_1 + 0x60);
                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                          }
                                          else {
                                            if ((-1 < (int)*plVar6) &&
                                               (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                              (**(code **)(plVar6[1] + 0x30))(plVar6);
                                            }
                                            plVar11 = (longlong *)FUN_18037f0c0();
                                            uVar4 = DAT_1803b32b0;
                                            plVar9 = DAT_1803b3288;
                                            plVar6 = DAT_1803b2d28;
                                            uVar15 = 0x213;
                                            if (plVar11 == (longlong *)0x0) {
                                              lVar8 = *(longlong *)(param_1 + 0x60);
                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                            }
                                            else {
                                              *(undefined4 *)(plVar3 + 5) = 0x213;
                                              local_68 = plVar6;
                                              local_60 = plVar9;
                                              local_50 = uVar4;
                                              local_58 = plVar10;
                                              plVar6 = (longlong *)
                                                       FUN_18037d280(param_1,plVar11,DAT_1803b32b8,
                                                                     &local_68,DAT_1803b32c0);
                                              if ((-1 < (int)*plVar11) &&
                                                 (*plVar11 = *plVar11 + -1, *plVar11 == 0)) {
                                                (**(code **)(plVar11[1] + 0x30))(plVar11);
                                              }
                                              if (plVar6 == (longlong *)0x0) {
                                                lVar8 = *(longlong *)(param_1 + 0x60);
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                              }
                                              else {
                                                if ((-1 < (int)*plVar6) &&
                                                   (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                  (**(code **)(plVar6[1] + 0x30))(plVar6);
                                                }
                                                uVar15 = DAT_1803b2d08;
                                                pcVar12 = (code *)_PyDict_NewPresized(5);
                                                PyDict_SetItem(pcVar12,uVar15,plVar10);
                                                PyDict_SetItem(pcVar12,DAT_1803b2d10,plVar7);
                                                PyDict_SetItem(pcVar12,DAT_1803b32c8,
                                                               _Py_TrueStruct_exref);
                                                PyDict_SetItem(pcVar12,DAT_1803b32d0,DAT_1803b3288);
                                                uVar15 = DAT_1803b2f68;
                                                plVar6 = (longlong *)
                                                         FUN_180382540(plVar1,DAT_1803b2c90);
                                                if (plVar6 != (longlong *)0x0) {
                                                  plVar7 = (longlong *)
                                                           FUN_180382540(plVar6,DAT_1803b2f70);
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    (**(code **)(plVar6[1] + 0x30))(plVar6);
                                                  }
                                                  if (plVar7 != (longlong *)0x0) {
                                                    plVar6 = (longlong *)FUN_1803749a0(plVar7);
                                                    if ((-1 < (int)*plVar7) &&
                                                       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                    }
                                                    if (plVar6 != (longlong *)0x0) {
                                                      PyDict_SetItem(pcVar12,uVar15,plVar6);
                                                      plVar7 = local_80;
                                                      if ((-1 < (int)*plVar6) &&
                                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                        (**(code **)(plVar6[1] + 0x30))(plVar6);
                                                        plVar7 = local_80;
                                                      }
                                                      goto LAB_1800d8606;
                                                    }
                                                  }
                                                }
                                                lVar8 = *(longlong *)(param_1 + 0x60);
                                                uVar15 = 0x21f;
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                if ((-1 < *(int *)pcVar12) &&
                                                   (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1
                                                   , *(longlong *)pcVar12 == 0)) {
                                                  (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))
                                                            (pcVar12);
                                                }
                                              }
                                            }
                                          }
                                        }
                                      }
                                    }
                                    goto LAB_1800d8366;
                                  }
                                }
                              }
                              lVar8 = *(longlong *)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              plVar10 = local_res20;
                              goto LAB_1800d8366;
                            }
                            uVar15 = 0x20c;
                          }
                        }
                      }
LAB_1800d7c88:
                      lVar8 = *(longlong *)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      plVar10 = local_res20;
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
  else {
    iVar5 = *(int *)(lVar14 + 0xc);
    if (iVar5 == 0) {
      iVar5 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
      *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar5 + 1;
      *(int *)(lVar14 + 0xc) = iVar5;
    }
    if (iVar5 != DAT_1803a9784) {
      DAT_1803a9784 = iVar5;
      DAT_1803b2988 = FUN_1803707d0(lVar14,DAT_1803b3278,*(undefined8 *)(DAT_1803b3278 + 0x18));
    }
    if (-1 < DAT_1803b2988) {
      lVar16 = (1L << (*(byte *)(lVar14 + 9) & 0x3f)) + lVar14;
      lVar8 = *(longlong *)(lVar16 + 0x28 + DAT_1803b2988 * 0x10);
      if (lVar8 != 0) goto LAB_1800d78a2;
      DAT_1803b2988 = FUN_1803707d0(lVar14,DAT_1803b3278,*(undefined8 *)(DAT_1803b3278 + 0x18));
      if (-1 < DAT_1803b2988) {
        lVar8 = *(longlong *)(lVar16 + 0x28 + DAT_1803b2988 * 0x10);
        goto LAB_1800d785c;
      }
    }
LAB_1800d7861:
    plVar6 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b3278);
    if ((plVar6 != (longlong *)0x0) && (lVar8 = *plVar6, lVar8 != 0)) goto LAB_1800d78a2;
    FUN_180373340(param_1,&local_90,DAT_1803b3278);
    uVar15 = 0x207;
    lVar8 = local_90;
  }
LAB_1800d8366:
  piVar2 = *(int **)(lVar8 + 0x28);
  if (piVar2 == (int *)0x0) {
    pcVar12 = (code *)FUN_18038e2a0(plVar3,uVar15);
    plVar6 = *(longlong **)(lVar8 + 0x28);
    pcVar13 = pcVar12;
    if (pcVar12 == _Py_NoneStruct_exref) {
      pcVar13 = (code *)0x0;
    }
    else {
LAB_1800d83c0:
      pcVar12 = pcVar13;
      if ((pcVar13 != (code *)0x0) && (*(int *)pcVar13 + 1 != 0)) {
        *(int *)pcVar13 = *(int *)pcVar13 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar2 + 6) == plVar3) goto LAB_1800d8409;
    pcVar12 = (code *)FUN_18038e2a0(plVar3,uVar15);
    *(int **)(pcVar12 + 0x10) = piVar2;
    if (*piVar2 + 1 != 0) {
      *piVar2 = *piVar2 + 1;
    }
    plVar6 = *(longlong **)(lVar8 + 0x28);
    pcVar13 = pcVar12;
    if (pcVar12 != _Py_NoneStruct_exref) goto LAB_1800d83c0;
    pcVar13 = (code *)0x0;
  }
  *(code **)(lVar8 + 0x28) = pcVar13;
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if (((pcVar12 != (code *)0x0) && (-1 < *(int *)pcVar12)) &&
     (*(longlong *)pcVar12 = *(longlong *)pcVar12 + -1, *(longlong *)pcVar12 == 0)) {
    (**(code **)(*(longlong *)(pcVar12 + 8) + 0x30))(pcVar12);
  }
LAB_1800d8409:
  plVar7 = local_48;
  FUN_18036d0a0(plVar3,"oooooooo",local_88,plVar1,plVar10,local_70,local_40,local_80,local_78,
                local_48);
  plVar6 = DAT_1803b3610;
  if (plVar3 == DAT_1803b3610) {
    if ((-1 < (int)*DAT_1803b3610) && (*DAT_1803b3610 = *DAT_1803b3610 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803b3610 = (longlong *)0x0;
  }
  lVar14 = **(longlong **)(param_1 + 0x38);
  plVar3 = *(longlong **)(lVar14 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar14 + 8);
  *(undefined4 *)(plVar3 + 8) = 0xffffffff;
  plVar6 = (longlong *)plVar3[2];
  if (((plVar6 != (longlong *)0x0) && (plVar3[2] = 0, -1 < (int)*plVar6)) &&
     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if ((-1 < (int)*plVar3) && (*plVar3 = *plVar3 + -1, *plVar3 == 0)) {
    (**(code **)(plVar3[1] + 0x30))(plVar3);
  }
  plVar3[10] = 0;
  if (((local_80 != (longlong *)0x0) && (-1 < (int)*local_80)) &&
     (*local_80 = *local_80 + -1, *local_80 == 0)) {
    (**(code **)(local_80[1] + 0x30))();
  }
  if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
     (*local_78 = *local_78 + -1, *local_78 == 0)) {
    (**(code **)(local_78[1] + 0x30))();
  }
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if ((-1 < (int)*local_88) && (*local_88 = *local_88 + -1, *local_88 == 0)) {
    (**(code **)(local_88[1] + 0x30))();
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
    (**(code **)(plVar10[1] + 0x30))(plVar10);
  }
  if ((-1 < (int)*local_70) && (*local_70 = *local_70 + -1, *local_70 == 0)) {
    (**(code **)(local_70[1] + 0x30))();
  }
  if ((-1 < (int)*local_40) && (*local_40 = *local_40 + -1, *local_40 == 0)) {
    (**(code **)(local_40[1] + 0x30))();
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar8;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (code *)0x0;
}
