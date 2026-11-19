/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180103dc0 */

/* WARNING: Function: __security_check_cookie replaced with injection: security_check_cookie */

undefined8 * FUN_180103dc0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  int *piVar1;
  longlong *plVar2;
  char cVar3;
  int iVar4;
  longlong *plVar5;
  longlong *plVar6;
  longlong *plVar7;
  longlong lVar8;
  longlong *plVar9;
  code *pcVar10;
  code *pcVar11;
  undefined8 *puVar12;
  longlong lVar13;
  undefined8 uVar14;
  longlong lVar15;
  longlong *plVar16;
  undefined1 auStack_248 [32];
  longlong *local_228;
  longlong *local_220;
  longlong *local_218;
  longlong *local_210;
  longlong *local_208;
  longlong *local_200;
  longlong *local_1f8;
  longlong *local_1f0;
  longlong *local_1e8;
  longlong *local_1e0;
  longlong *local_1d8;
  longlong *local_1d0;
  longlong *local_1c8;
  longlong *local_1c0;
  longlong *local_1b8;
  longlong *local_1a8;
  longlong *local_1a0;
  longlong *local_198;
  longlong *local_190;
  longlong *local_188;
  longlong *local_180;
  longlong *local_170;
  longlong *local_168;
  longlong *local_160;
  longlong *local_158;
  longlong *local_150;
  longlong *local_148;
  longlong *local_140;
  longlong *local_138;
  longlong *local_130;
  longlong *local_128;
  longlong *local_120;
  longlong *local_118;
  longlong *local_110;
  longlong *local_108;
  longlong *local_100;
  longlong *local_f8;
  longlong *local_f0;
  longlong *local_e8;
  longlong *local_e0;
  longlong *local_d8;
  longlong *local_d0;
  longlong *local_c8;
  longlong *local_c0;
  longlong *local_b8;
  longlong *local_b0;
  longlong *local_a8;
  longlong *local_a0;
  longlong *local_98;
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;
  ulonglong local_40;
  
  plVar9 = DAT_1803b3958;
  local_40 = DAT_1803a6040 ^ (ulonglong)auStack_248;
  local_110 = (longlong *)param_3[2];
  plVar7 = (longlong *)*param_3;
  plVar16 = (longlong *)0x0;
  plVar6 = (longlong *)param_3[1];
  local_150 = (longlong *)param_3[3];
  local_138 = (longlong *)param_3[4];
  local_130 = (longlong *)param_3[5];
  local_128 = (longlong *)param_3[6];
  local_170 = DAT_180391d10;
  local_160 = (longlong *)0x0;
  local_148 = (longlong *)0x0;
  local_108 = (longlong *)0x0;
  local_168 = (longlong *)0x0;
  local_100 = (longlong *)0x0;
  local_e8 = (longlong *)0x0;
  local_120 = (longlong *)0x0;
  local_158 = (longlong *)0x0;
  local_1a0 = (longlong *)0x0;
  local_118 = (longlong *)0x0;
  local_188 = (longlong *)0x0;
  local_190 = DAT_1803b3958;
  local_1a8 = plVar7;
  local_140 = plVar6;
  if (DAT_1803b3958 == (longlong *)0x0) {
LAB_180103ea2:
    DAT_1803b3958 = (longlong *)FUN_18036cd10(param_1,DAT_1803b3e88,DAT_1803ca778,0x88);
  }
  else if ((1 < *DAT_1803b3958) || (DAT_1803b3958[2] != 0)) {
    if ((DAT_1803b3958 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803b3958 && (*DAT_1803b3958 = *DAT_1803b3958 + -1, *plVar9 == 0)))) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
    goto LAB_180103ea2;
  }
  plVar9 = DAT_1803b3958;
  local_190 = DAT_1803b3958;
  FUN_180001d00(param_1,DAT_1803b3958);
  plVar5 = (longlong *)FUN_18037f0c0();
  if (plVar5 == (longlong *)0x0) {
    plVar6 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    uVar14 = 0x355;
    plVar16 = (longlong *)0x0;
  }
  else {
    uVar14 = 0x355;
    *(undefined4 *)(plVar9 + 5) = 0x355;
    plVar6 = (longlong *)FUN_18037e6c0(param_1,plVar5,DAT_1803b41c8,plVar6);
    local_e0 = plVar6;
    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    if (plVar6 == (longlong *)0x0) {
      plVar6 = *(longlong **)(param_1 + 0x60);
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      lVar8 = *(longlong *)(DAT_1803ca780 + 0x20);
      local_160 = plVar6;
      if (*(char *)(lVar8 + 10) == '\0') {
        plVar5 = (longlong *)FUN_1800015a0();
        if (plVar5 == (longlong *)0x0) {
          lVar13 = 0;
        }
        else {
          lVar13 = *plVar5;
        }
LAB_18010401f:
        if (lVar13 == 0) goto LAB_180104024;
      }
      else {
        iVar4 = *(int *)(lVar8 + 0xc);
        if (iVar4 == 0) {
          iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
          *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
          *(int *)(lVar8 + 0xc) = iVar4;
        }
        if (iVar4 != DAT_1803a9844) {
          DAT_1803a9844 = iVar4;
          DAT_1803b3ef0 = FUN_1803707d0(lVar8,DAT_1803b4b80,*(undefined8 *)(DAT_1803b4b80 + 0x18));
        }
        if (-1 < DAT_1803b3ef0) {
          lVar15 = (1L << (*(byte *)(lVar8 + 9) & 0x3f)) + lVar8;
          lVar13 = *(longlong *)(lVar15 + 0x28 + DAT_1803b3ef0 * 0x10);
          if (lVar13 != 0) goto LAB_180104065;
          DAT_1803b3ef0 = FUN_1803707d0(lVar8,DAT_1803b4b80,*(undefined8 *)(DAT_1803b4b80 + 0x18));
          if (-1 < DAT_1803b3ef0) {
            lVar13 = *(longlong *)(lVar15 + 0x28 + DAT_1803b3ef0 * 0x10);
            goto LAB_18010401f;
          }
        }
LAB_180104024:
        plVar5 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b4b80);
        if ((plVar5 == (longlong *)0x0) || (lVar13 = *plVar5, lVar13 == 0)) {
          FUN_180373340(param_1,&local_170,DAT_1803b4b80);
          uVar14 = 0x356;
          plVar6 = local_170;
          goto LAB_18010450f;
        }
      }
LAB_180104065:
      uVar14 = 0x356;
      *(undefined4 *)(plVar9 + 5) = 0x356;
      local_f0 = (longlong *)FUN_180376770(param_1,lVar13,plVar6);
      if (local_f0 == (longlong *)0x0) {
        plVar6 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
        lVar8 = *(longlong *)(DAT_1803ca780 + 0x20);
        local_148 = local_f0;
        if (*(char *)(lVar8 + 10) == '\0') {
          plVar7 = (longlong *)FUN_1800015a0(DAT_1803ca780,DAT_1803b4b88);
          if (plVar7 == (longlong *)0x0) {
            lVar13 = 0;
          }
          else {
            lVar13 = *plVar7;
          }
LAB_180104160:
          if (lVar13 == 0) goto LAB_180104165;
        }
        else {
          iVar4 = *(int *)(lVar8 + 0xc);
          if (iVar4 == 0) {
            iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
            *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
            *(int *)(lVar8 + 0xc) = iVar4;
          }
          if (iVar4 != DAT_1803a9878) {
            DAT_1803a9878 = iVar4;
            DAT_1803b3d30 = FUN_1803707d0(lVar8,DAT_1803b4b88,*(undefined8 *)(DAT_1803b4b88 + 0x18))
            ;
          }
          if (-1 < DAT_1803b3d30) {
            lVar15 = (1L << (*(byte *)(lVar8 + 9) & 0x3f)) + lVar8;
            lVar13 = *(longlong *)(lVar15 + 0x28 + DAT_1803b3d30 * 0x10);
            if (lVar13 != 0) goto LAB_1801041ab;
            DAT_1803b3d30 = FUN_1803707d0(lVar8,DAT_1803b4b88,*(undefined8 *)(DAT_1803b4b88 + 0x18))
            ;
            if (-1 < DAT_1803b3d30) {
              lVar13 = *(longlong *)(lVar15 + 0x28 + DAT_1803b3d30 * 0x10);
              goto LAB_180104160;
            }
          }
LAB_180104165:
          plVar7 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b4b88);
          if (plVar7 == (longlong *)0x0) {
            lVar13 = 0;
          }
          else {
            lVar13 = *plVar7;
          }
          if (lVar13 == 0) {
            FUN_180373340(param_1,&local_170,DAT_1803b4b88);
            uVar14 = 0x357;
            plVar6 = local_170;
            plVar16 = (longlong *)0x0;
            plVar7 = local_1a8;
            goto LAB_18010450f;
          }
        }
LAB_1801041ab:
        plVar7 = (longlong *)FUN_18037f0c0();
        if (plVar7 == (longlong *)0x0) {
          plVar6 = *(longlong **)(param_1 + 0x60);
          uVar14 = 0x357;
          *(undefined8 *)(param_1 + 0x60) = 0;
          plVar16 = local_1a0;
          plVar7 = local_1a8;
        }
        else {
          lVar8 = *(longlong *)(param_1 + 0x10);
          plVar6 = *(longlong **)(lVar8 + 0x41da0);
          if (plVar6 == (longlong *)0x0) {
            plVar6 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,2);
          }
          else {
            *(longlong *)(lVar8 + 0x41da0) = plVar6[3];
            *(int *)(lVar8 + 0x41e3c) = *(int *)(lVar8 + 0x41e3c) + -1;
            *plVar6 = 1;
          }
          plVar9 = plVar6 + -2;
          plVar6[3] = 0;
          plVar6[4] = 0;
          lVar8 = _PyThreadState_GetCurrent();
          lVar8 = *(longlong *)(*(longlong *)(lVar8 + 0x10) + 0xd0);
          puVar12 = *(undefined8 **)(lVar8 + 8);
          *puVar12 = plVar9;
          plVar6[-1] = plVar6[-1] & 3;
          plVar6[-1] = plVar6[-1] | (ulonglong)puVar12;
          *plVar9 = lVar8;
          *(longlong **)(lVar8 + 8) = plVar9;
          iVar4 = (int)*local_140 + 1;
          if (iVar4 != 0) {
            *(int *)local_140 = iVar4;
          }
          plVar6[3] = (longlong)local_140;
          iVar4 = (int)*local_128 + 1;
          if (iVar4 != 0) {
            *(int *)local_128 = iVar4;
          }
          plVar6[4] = (longlong)local_128;
          plVar9 = (longlong *)FUN_18036fb80(param_1,DAT_1803b4b90);
          uVar14 = 0x357;
          *(undefined4 *)(local_190 + 5) = 0x357;
          plVar16 = (longlong *)FUN_180001790(param_1,plVar7,plVar6,plVar9);
          if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
            (**(code **)(plVar7[1] + 0x30))(plVar7);
          }
          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
            (**(code **)(plVar6[1] + 0x30))(plVar6);
          }
          if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
            (**(code **)(plVar9[1] + 0x30))(plVar9);
          }
          plVar9 = local_190;
          if (plVar16 == (longlong *)0x0) {
            plVar6 = *(longlong **)(param_1 + 0x60);
            *(undefined8 *)(param_1 + 0x60) = 0;
            plVar16 = local_1a0;
            plVar7 = local_1a8;
          }
          else {
            *(undefined4 *)(local_190 + 5) = 0x357;
            local_f8 = (longlong *)FUN_180376770(param_1,lVar13,plVar16);
            if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
              (**(code **)(plVar16[1] + 0x30))(plVar16);
            }
            if (local_f8 == (longlong *)0x0) {
              plVar6 = *(longlong **)(param_1 + 0x60);
              *(undefined8 *)(param_1 + 0x60) = 0;
              plVar16 = local_1a0;
              plVar7 = local_1a8;
            }
            else {
              local_68 = local_110;
              uVar14 = 0x358;
              local_60 = local_160;
              local_58 = local_150;
              local_50 = local_138;
              local_48 = local_130;
              *(undefined4 *)(plVar9 + 5) = 0x358;
              local_108 = local_f8;
              local_90 = (longlong *)FUN_18037ec40(param_1,local_f8,DAT_1803b4b98,&local_68);
              if (local_90 == (longlong *)0x0) {
                plVar6 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar16 = local_1a0;
                plVar7 = local_1a8;
              }
              else {
                lVar8 = *(longlong *)(DAT_1803ca780 + 0x20);
                local_168 = local_90;
                if (*(char *)(lVar8 + 10) == '\0') {
                  plVar7 = (longlong *)FUN_1800015a0(DAT_1803ca780,DAT_1803b4ba0);
                  if (plVar7 == (longlong *)0x0) {
                    lVar13 = 0;
                  }
                  else {
                    lVar13 = *plVar7;
                  }
LAB_180104492:
                  if (lVar13 == 0) goto LAB_180104497;
                }
                else {
                  iVar4 = *(int *)(lVar8 + 0xc);
                  if (iVar4 == 0) {
                    iVar4 = *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118);
                    *(int *)(*(longlong *)(param_1 + 0x10) + 0x42118) = iVar4 + 1;
                    *(int *)(lVar8 + 0xc) = iVar4;
                  }
                  if (iVar4 != DAT_1803a9864) {
                    DAT_1803a9864 = iVar4;
                    DAT_1803b3c70 =
                         FUN_1803707d0(lVar8,DAT_1803b4ba0,*(undefined8 *)(DAT_1803b4ba0 + 0x18));
                  }
                  if (-1 < DAT_1803b3c70) {
                    lVar15 = (1L << (*(byte *)(lVar8 + 9) & 0x3f)) + lVar8;
                    lVar13 = *(longlong *)(lVar15 + 0x28 + DAT_1803b3c70 * 0x10);
                    if (lVar13 != 0) goto LAB_1801044da;
                    DAT_1803b3c70 =
                         FUN_1803707d0(lVar8,DAT_1803b4ba0,*(undefined8 *)(DAT_1803b4ba0 + 0x18));
                    if (-1 < DAT_1803b3c70) {
                      lVar13 = *(longlong *)(lVar15 + 0x28 + DAT_1803b3c70 * 0x10);
                      goto LAB_180104492;
                    }
                  }
LAB_180104497:
                  plVar7 = (longlong *)FUN_1800015a0(DAT_1803ab778,DAT_1803b4ba0);
                  if (plVar7 == (longlong *)0x0) {
                    lVar13 = 0;
                  }
                  else {
                    lVar13 = *plVar7;
                  }
                  if (lVar13 == 0) {
                    FUN_180373340(param_1,&local_170,DAT_1803b4ba0);
                    uVar14 = 0x359;
                    plVar6 = local_170;
                    plVar16 = local_1a0;
                    plVar9 = local_190;
                    plVar7 = local_1a8;
                    goto LAB_18010450f;
                  }
                }
LAB_1801044da:
                plVar7 = (longlong *)FUN_18037f0c0();
                if (plVar7 == (longlong *)0x0) {
                  uVar14 = 0x359;
                }
                else {
                  lVar8 = *(longlong *)(param_1 + 0x10);
                  plVar6 = *(longlong **)(lVar8 + 0x41da0);
                  if (plVar6 == (longlong *)0x0) {
                    plVar6 = (longlong *)FUN_18035ebd0(PyTuple_Type_exref,2);
                  }
                  else {
                    *(longlong *)(lVar8 + 0x41da0) = plVar6[3];
                    *(int *)(lVar8 + 0x41e3c) = *(int *)(lVar8 + 0x41e3c) + -1;
                    *plVar6 = 1;
                  }
                  plVar9 = plVar6 + -2;
                  plVar6[3] = 0;
                  plVar6[4] = 0;
                  lVar8 = _PyThreadState_GetCurrent();
                  lVar8 = *(longlong *)(*(longlong *)(lVar8 + 0x10) + 0xd0);
                  puVar12 = *(undefined8 **)(lVar8 + 8);
                  *puVar12 = plVar9;
                  plVar6[-1] = plVar6[-1] & 3;
                  plVar6[-1] = plVar6[-1] | (ulonglong)puVar12;
                  *plVar9 = lVar8;
                  *(longlong **)(lVar8 + 8) = plVar9;
                  iVar4 = (int)*local_140 + 1;
                  if (iVar4 != 0) {
                    *(int *)local_140 = iVar4;
                  }
                  plVar6[3] = (longlong)local_140;
                  iVar4 = (int)*local_128 + 1;
                  if (iVar4 != 0) {
                    *(int *)local_128 = iVar4;
                  }
                  plVar6[4] = (longlong)local_128;
                  plVar9 = (longlong *)FUN_18036fb80(param_1,DAT_1803b4ba8);
                  uVar14 = 0x359;
                  *(undefined4 *)(local_190 + 5) = 0x359;
                  plVar16 = (longlong *)FUN_180001790(param_1,plVar7,plVar6,plVar9);
                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                  }
                  if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                    (**(code **)(plVar6[1] + 0x30))(plVar6);
                  }
                  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                  }
                  plVar9 = local_190;
                  if (plVar16 != (longlong *)0x0) {
                    *(undefined4 *)(local_190 + 5) = 0x359;
                    plVar7 = (longlong *)FUN_180376770(param_1,lVar13,plVar16);
                    local_d8 = plVar7;
                    if ((-1 < (int)*plVar16) && (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                      (**(code **)(plVar16[1] + 0x30))(plVar16);
                    }
                    if (plVar7 == (longlong *)0x0) {
                      plVar6 = *(longlong **)(param_1 + 0x60);
                      *(undefined8 *)(param_1 + 0x60) = 0;
                      plVar16 = local_1a0;
                      plVar7 = local_1a8;
                    }
                    else {
                      local_100 = plVar7;
                      lVar8 = FUN_1800df1c0(param_1);
                      if (lVar8 == 0) {
                        FUN_180373340(param_1,&local_170,DAT_1803b41b8);
                        uVar14 = 0x35a;
                        plVar6 = local_170;
                        plVar16 = local_1a0;
                        plVar7 = local_1a8;
                      }
                      else {
                        plVar7 = (longlong *)FUN_18037f0c0();
                        if (plVar7 == (longlong *)0x0) {
                          plVar6 = *(longlong **)(param_1 + 0x60);
                          uVar14 = 0x35a;
                          *(undefined8 *)(param_1 + 0x60) = 0;
                          plVar16 = local_1a0;
                          plVar7 = local_1a8;
                        }
                        else {
                          plVar6 = (longlong *)FUN_180382540(local_160,DAT_1803b4bb0);
                          if (plVar6 == (longlong *)0x0) {
                            plVar6 = *(longlong **)(param_1 + 0x60);
                            *(undefined8 *)(param_1 + 0x60) = 0;
                            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                            }
                            uVar14 = 0x35a;
                            plVar16 = local_1a0;
                            plVar7 = local_1a8;
                          }
                          else {
                            uVar14 = 0x35a;
                            local_80 = DAT_1803b3f30;
                            *(undefined4 *)(plVar9 + 5) = 0x35a;
                            local_88 = plVar6;
                            plVar16 = (longlong *)FUN_180377750(param_1,plVar7,&local_88);
                            local_d0 = plVar16;
                            if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                            }
                            if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                              (**(code **)(plVar6[1] + 0x30))(plVar6);
                            }
                            if (plVar16 == (longlong *)0x0) {
                              plVar6 = *(longlong **)(param_1 + 0x60);
                              *(undefined8 *)(param_1 + 0x60) = 0;
                              plVar16 = local_1a0;
                              plVar7 = local_1a8;
                            }
                            else {
                              local_e8 = plVar16;
                              lVar8 = FUN_1800df1c0(param_1);
                              if (lVar8 == 0) {
                                FUN_180373340(param_1,&local_170,DAT_1803b41b8);
                                uVar14 = 0x35b;
                                plVar6 = local_170;
                                plVar16 = local_1a0;
                                plVar7 = local_1a8;
                              }
                              else {
                                plVar7 = (longlong *)FUN_18037f0c0();
                                if (plVar7 == (longlong *)0x0) {
                                  plVar6 = *(longlong **)(param_1 + 0x60);
                                  uVar14 = 0x35b;
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  plVar16 = local_1a0;
                                  plVar7 = local_1a8;
                                }
                                else {
                                  plVar6 = (longlong *)FUN_180382540(local_160,DAT_1803b4bb8);
                                  if (plVar6 != (longlong *)0x0) {
                                    uVar14 = 0x35b;
                                    local_70 = DAT_1803b3f30;
                                    *(undefined4 *)(plVar9 + 5) = 0x35b;
                                    local_78 = plVar6;
                                    plVar9 = (longlong *)FUN_180377750(param_1,plVar7,&local_78);
                                    local_198 = plVar9;
                                    if ((-1 < (int)*plVar7) &&
                                       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                      (**(code **)(plVar7[1] + 0x30))(plVar7);
                                    }
                                    if ((-1 < (int)*plVar6) &&
                                       (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                      (**(code **)(plVar6[1] + 0x30))(plVar6);
                                    }
                                    if (plVar9 != (longlong *)0x0) {
                                      local_120 = plVar9;
                                      plVar7 = (longlong *)FUN_18037f0c0();
                                      if (plVar7 != (longlong *)0x0) {
                                        plVar5 = (longlong *)FUN_18037f0c0();
                                        if (plVar5 == (longlong *)0x0) {
                                          plVar6 = *(longlong **)(param_1 + 0x60);
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                          if ((-1 < (int)*plVar7) &&
                                             (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                            (**(code **)(plVar7[1] + 0x30))(plVar7);
                                          }
                                          uVar14 = 0x360;
                                          plVar16 = local_1a0;
                                          plVar9 = local_190;
                                          plVar7 = local_1a8;
                                        }
                                        else {
                                          plVar6 = (longlong *)FUN_18037f0c0();
                                          if (plVar6 != (longlong *)0x0) {
                                            uVar14 = 0x35c;
                                            *(undefined4 *)(local_190 + 5) = 0x35c;
                                            local_b8 = local_150;
                                            local_b0 = local_130;
                                            local_a8 = local_148;
                                            local_228 = (longlong *)DAT_1803b4bd8;
                                            local_c8 = plVar16;
                                            local_c0 = plVar9;
                                            local_a0 = plVar5;
                                            local_98 = plVar6;
                                            plVar9 = (longlong *)
                                                     FUN_18037dfa0(param_1,plVar7,&local_b8,
                                                                   &local_c8);
                                            local_c8 = plVar9;
                                            if ((-1 < (int)*plVar7) &&
                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                              (**(code **)(plVar7[1] + 0x30))(plVar7);
                                            }
                                            if ((-1 < (int)*plVar5) &&
                                               (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                              (**(code **)(plVar5[1] + 0x30))(plVar5);
                                            }
                                            if ((-1 < (int)*plVar6) &&
                                               (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                              (**(code **)(plVar6[1] + 0x30))(plVar6);
                                            }
                                            if (plVar9 == (longlong *)0x0) goto LAB_1801044f7;
                                            local_158 = plVar9;
                                            plVar7 = (longlong *)FUN_18037f0c0();
                                            if (plVar7 != (longlong *)0x0) {
                                              plVar5 = (longlong *)FUN_180001930();
                                              if ((-1 < (int)*plVar7) &&
                                                 (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                (**(code **)(plVar7[1] + 0x30))(plVar7);
                                              }
                                              if (plVar5 != (longlong *)0x0) {
                                                plVar6 = (longlong *)FUN_18000d430(plVar5);
                                                if (plVar6 != (longlong *)0x0) {
                                                  uVar14 = 0x366;
                                                  do {
                                                    plVar7 = local_1a0;
                                                    plVar9 = plVar6;
                                                    if (((local_188 != (longlong *)0x0) &&
                                                        (-1 < (int)*local_188)) &&
                                                       (*local_188 = *local_188 + -1,
                                                       *local_188 == 0)) {
                                                      plVar9 = local_188 + 1;
                                                      local_188 = plVar6;
                                                      (**(code **)(*plVar9 + 0x30))();
                                                      plVar9 = local_188;
                                                    }
                                                    local_188 = plVar9;
                                                    iVar4 = (int)*plVar6 + 1;
                                                    if (iVar4 != 0) {
                                                      *(int *)plVar6 = iVar4;
                                                    }
                                                    local_1a0 = plVar6;
                                                    if (((plVar7 != (longlong *)0x0) &&
                                                        (-1 < (int)*plVar7)) &&
                                                       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                      (**(code **)(plVar7[1] + 0x30))();
                                                    }
                                                    plVar7 = local_1a8;
                                                    if (local_1a8 == (longlong *)0x0) {
                                                      plVar9 = (longlong *)
                                                               PyUnicode_FromFormat
                                                                         (
                                                  "cannot access local variable \'%U\' where it is not associated with a value"
                                                  ,DAT_1803b3fb8);
                                                  plVar6 = (longlong *)
                                                           FUN_180376770(param_1,*(undefined8 *)
                                                                                                                                                                    
                                                  PyExc_UnboundLocalError_exref,plVar9);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                  }
                                                  FUN_18000d1d0(param_1,plVar6);
                                                  goto LAB_180104cf7;
                                                  }
                                                  plVar9 = (longlong *)FUN_18037f0c0();
                                                  piVar1 = DAT_1803b4be0;
                                                  if (plVar9 == (longlong *)0x0) {
LAB_180104d40:
                                                    plVar6 = *(longlong **)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    plVar7 = local_1a8;
                                                    goto LAB_180104cf7;
                                                  }
                                                  lVar8 = *(longlong *)(param_1 + 0x10);
                                                  plVar16 = *(longlong **)(lVar8 + 0x41da0);
                                                  if (plVar16 == (longlong *)0x0) {
                                                    plVar16 = (longlong *)
                                                              FUN_18035ebd0(PyTuple_Type_exref,2);
                                                  }
                                                  else {
                                                    *(longlong *)(lVar8 + 0x41da0) = plVar16[3];
                                                    *(int *)(lVar8 + 0x41e3c) =
                                                         *(int *)(lVar8 + 0x41e3c) + -1;
                                                    *plVar16 = 1;
                                                  }
                                                  plVar16[3] = 0;
                                                  plVar16[4] = 0;
                                                  FUN_1801f33c0(plVar16);
                                                  if (*piVar1 + 1 != 0) {
                                                    *piVar1 = *piVar1 + 1;
                                                  }
                                                  plVar16[3] = (longlong)piVar1;
                                                  lVar8 = FUN_180375520(param_1,plVar6,DAT_1803b3f30
                                                                       );
                                                  if (lVar8 == 0) {
                                                    plVar6 = *(longlong **)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                    }
                                                    plVar7 = local_1a8;
                                                    if ((-1 < (int)*plVar16) &&
                                                       (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                      (**(code **)(plVar16[1] + 0x30))(plVar16);
                                                      plVar7 = local_1a8;
                                                    }
                                                    goto LAB_180104cf7;
                                                  }
                                                  plVar16[4] = lVar8;
                                                  plVar7 = (longlong *)
                                                           PyUnicode_Join(DAT_1803b3f30,plVar16);
                                                  if ((-1 < (int)*plVar16) &&
                                                     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                    (**(code **)(plVar16[1] + 0x30))(plVar16);
                                                  }
                                                  if (plVar7 == (longlong *)0x0) {
                                                    plVar6 = *(longlong **)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    plVar7 = local_1a8;
                                                    if ((-1 < (int)*plVar9) &&
                                                       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                      (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                      plVar7 = local_1a8;
                                                    }
                                                    goto LAB_180104cf7;
                                                  }
                                                  *(undefined4 *)(local_190 + 5) = 0x366;
                                                  plVar6 = (longlong *)
                                                           FUN_180376770(param_1,plVar9,plVar7);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                  }
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                  }
                                                  if (plVar6 == (longlong *)0x0) goto LAB_180104d40;
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    (**(code **)(plVar6[1] + 0x30))(plVar6);
                                                  }
                                                  cVar3 = FUN_18000d5b0(param_1);
                                                  if (cVar3 == '\0') {
                                                    plVar6 = *(longlong **)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    goto LAB_180104cee;
                                                  }
                                                  plVar6 = (longlong *)FUN_18000d430(plVar5);
                                                  } while (plVar6 != (longlong *)0x0);
                                                }
                                                plVar6 = *(longlong **)(param_1 + 0x60);
                                                if (plVar6 != (longlong *)0x0) {
                                                  cVar3 = FUN_1800014b0();
                                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                                  if (cVar3 == '\0') {
LAB_180104cee:
                                                    uVar14 = 0x365;
                                                    plVar7 = local_1a8;
LAB_180104cf7:
                                                    plVar16 = local_1a0;
                                                    if (((local_1a0 != (longlong *)0x0) &&
                                                        (-1 < (int)*local_1a0)) &&
                                                       (*local_1a0 = *local_1a0 + -1,
                                                       *local_1a0 == 0)) {
                                                      (**(code **)(local_1a0[1] + 0x30))(local_1a0);
                                                    }
                                                    plVar9 = local_190;
                                                    if ((-1 < (int)*plVar5) &&
                                                       (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                      (**(code **)(plVar5[1] + 0x30))(plVar5);
                                                      plVar9 = local_190;
                                                    }
                                                    goto LAB_18010450f;
                                                  }
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    (**(code **)(plVar6[1] + 0x30))(plVar6);
                                                  }
                                                }
                                                plVar16 = local_1a0;
                                                if (((local_1a0 != (longlong *)0x0) &&
                                                    (-1 < (int)*local_1a0)) &&
                                                   (*local_1a0 = *local_1a0 + -1, *local_1a0 == 0))
                                                {
                                                  (**(code **)(local_1a0[1] + 0x30))(local_1a0);
                                                }
                                                if ((-1 < (int)*plVar5) &&
                                                   (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                  (**(code **)(plVar5[1] + 0x30))(plVar5);
                                                }
                                                plVar7 = (longlong *)FUN_18037f0c0();
                                                if (plVar7 != (longlong *)0x0) {
                                                  iVar4 = FUN_180001510(plVar7);
                                                  if ((-1 < (int)*plVar7) &&
                                                     (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                  }
                                                  plVar9 = local_190;
                                                  plVar7 = local_1a8;
                                                  if (iVar4 != -1) {
                                                    if (iVar4 == 0) {
                                                      if (local_1a8 == (longlong *)0x0) {
                                                        plVar9 = (longlong *)
                                                                 PyUnicode_FromFormat
                                                                           (
                                                  "cannot access local variable \'%U\' where it is not associated with a value"
                                                  ,DAT_1803b3fb8);
                                                  plVar6 = (longlong *)
                                                           FUN_180376770(param_1,*(undefined8 *)
                                                                                                                                                                    
                                                  PyExc_UnboundLocalError_exref,plVar9);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                  }
                                                  FUN_18000d1d0(param_1,plVar6);
                                                  uVar14 = 0x368;
                                                  plVar9 = local_190;
                                                  }
                                                  else {
                                                    uVar14 = 0x368;
                                                    *(undefined4 *)(local_190 + 5) = 0x368;
                                                    plVar6 = (longlong *)
                                                             FUN_18037e6c0(param_1,local_1a8,
                                                                           DAT_1803b4260,
                                                                           *(undefined8 *)
                                                                            (DAT_1803b4be8 + 0x18));
                                                    if (plVar6 != (longlong *)0x0) {
                                                      if ((-1 < (int)*plVar6) &&
                                                         (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                        (**(code **)(plVar6[1] + 0x30))();
                                                      }
                                                      goto LAB_180104f7d;
                                                    }
                                                    plVar6 = *(longlong **)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                  }
                                                  }
                                                  else {
LAB_180104f7d:
                                                    lVar8 = FUN_1800e1240(param_1);
                                                    if (lVar8 == 0) {
                                                      FUN_180373340(param_1,&local_170,DAT_1803b4bf0
                                                                   );
                                                      uVar14 = 0x369;
                                                      plVar6 = local_170;
                                                      plVar16 = local_1a0;
                                                    }
                                                    else {
                                                      plVar16 = (longlong *)FUN_18037f0c0();
                                                      if (plVar16 == (longlong *)0x0) {
                                                        plVar6 = *(longlong **)(param_1 + 0x60);
                                                        uVar14 = 0x36b;
                                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                                        plVar16 = local_1a0;
                                                      }
                                                      else {
                                                        plVar7 = (longlong *)FUN_18037f0c0();
                                                        if (plVar7 == (longlong *)0x0) {
                                                          plVar6 = *(longlong **)(param_1 + 0x60);
                                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                                          if ((-1 < (int)*plVar16) &&
                                                             (*plVar16 = *plVar16 + -1,
                                                             *plVar16 == 0)) {
                                                            (**(code **)(plVar16[1] + 0x30))
                                                                      (plVar16);
                                                          }
                                                          uVar14 = 0x36c;
                                                          plVar16 = local_1a0;
                                                          plVar7 = local_1a8;
                                                        }
                                                        else {
                                                          plVar6 = (longlong *)FUN_18037f0c0();
                                                          if (plVar6 == (longlong *)0x0) {
                                                            plVar6 = *(longlong **)(param_1 + 0x60);
                                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                                            if ((-1 < (int)*plVar16) &&
                                                               (*plVar16 = *plVar16 + -1,
                                                               *plVar16 == 0)) {
                                                              (**(code **)(plVar16[1] + 0x30))
                                                                        (plVar16);
                                                            }
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              (**(code **)(plVar7[1] + 0x30))
                                                                        (plVar7);
                                                            }
                                                            uVar14 = 0x36d;
                                                            plVar16 = local_1a0;
                                                            plVar7 = local_1a8;
                                                          }
                                                          else {
                                                            local_188 = local_160;
                                                            uVar14 = 0x369;
                                                            *(undefined4 *)(plVar9 + 5) = 0x369;
                                                            local_228 = (longlong *)DAT_1803b4c00;
                                                            local_b8 = plVar16;
                                                            local_b0 = plVar7;
                                                            local_a8 = plVar6;
                                                            plVar5 = (longlong *)
                                                                     FUN_18037d030(param_1,lVar8,
                                                                                   &local_188,
                                                                                   &local_b8);
                                                            if ((-1 < (int)*plVar16) &&
                                                               (*plVar16 = *plVar16 + -1,
                                                               *plVar16 == 0)) {
                                                              (**(code **)(plVar16[1] + 0x30))
                                                                        (plVar16);
                                                            }
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              (**(code **)(plVar7[1] + 0x30))
                                                                        (plVar7);
                                                            }
                                                            if ((-1 < (int)*plVar6) &&
                                                               (*plVar6 = *plVar6 + -1, *plVar6 == 0
                                                               )) {
                                                              (**(code **)(plVar6[1] + 0x30))
                                                                        (plVar6);
                                                            }
                                                            if (plVar5 == (longlong *)0x0) {
                                                              plVar6 = *(longlong **)
                                                                        (param_1 + 0x60);
                                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                                              plVar16 = local_1a0;
                                                              plVar7 = local_1a8;
                                                            }
                                                            else {
                                                              if ((-1 < (int)*plVar5) &&
                                                                 (*plVar5 = *plVar5 + -1,
                                                                 *plVar5 == 0)) {
                                                                (**(code **)(plVar5[1] + 0x30))
                                                                          (plVar5);
                                                              }
                                                              lVar8 = FUN_1800e1130(param_1);
                                                              if (lVar8 == 0) {
                                                                FUN_180373340(param_1,&local_170,
                                                                              DAT_1803b4c08);
                                                                uVar14 = 0x36f;
                                                                plVar6 = local_170;
                                                                plVar16 = local_1a0;
                                                                plVar7 = local_1a8;
                                                              }
                                                              else {
                                                                plVar7 = (longlong *)FUN_18037f0c0()
                                                                ;
                                                                if (plVar7 == (longlong *)0x0) {
                                                                  plVar6 = *(longlong **)
                                                                            (param_1 + 0x60);
                                                                  uVar14 = 0x373;
                                                                  *(undefined8 *)(param_1 + 0x60) =
                                                                       0;
                                                                  plVar16 = local_1a0;
                                                                  plVar7 = local_1a8;
                                                                }
                                                                else {
                                                                  iVar4 = FUN_180001510(plVar7);
                                                                  if (iVar4 == -1) {
                                                                    plVar6 = *(longlong **)
                                                                              (param_1 + 0x60);
                                                                    *(undefined8 *)(param_1 + 0x60)
                                                                         = 0;
                                                                    if ((-1 < (int)*plVar7) &&
                                                                       (*plVar7 = *plVar7 + -1,
                                                                       *plVar7 == 0)) {
                                                                      (**(code **)(plVar7[1] + 0x30)
                                                                      )(plVar7);
                                                                    }
                                                                    uVar14 = 0x373;
                                                                    plVar16 = local_1a0;
                                                                    plVar7 = local_1a8;
                                                                  }
                                                                  else {
                                                                    if (iVar4 != 1) {
                                                                      if ((-1 < (int)*plVar7) &&
                                                                         (*plVar7 = *plVar7 + -1,
                                                                         *plVar7 == 0)) {
                                                                        (**(code **)(plVar7[1] +
                                                                                    0x30))(plVar7);
                                                                      }
                                                                      plVar7 = DAT_1803b3f30;
                                                                      iVar4 = (int)*DAT_1803b3f30 +
                                                                              1;
                                                                      if (iVar4 != 0) {
                                                                        *(int *)DAT_1803b3f30 =
                                                                             iVar4;
                                                                      }
                                                                    }
                                                                    plVar6 = (longlong *)
                                                                             FUN_18037f0c0();
                                                                    if ((-1 < (int)*plVar7) &&
                                                                       (*plVar7 = *plVar7 + -1,
                                                                       *plVar7 == 0)) {
                                                                      (**(code **)(plVar7[1] + 0x30)
                                                                      )(plVar7);
                                                                    }
                                                                    uVar14 = 0x373;
                                                                    if (plVar6 == (longlong *)0x0) {
                                                                      plVar6 = *(longlong **)
                                                                                (param_1 + 0x60);
                                                                      *(undefined8 *)
                                                                       (param_1 + 0x60) = 0;
                                                                      plVar16 = local_1a0;
                                                                      plVar7 = local_1a8;
                                                                    }
                                                                    else {
                                                                      *(undefined4 *)(plVar9 + 5) =
                                                                           0x373;
                                                                      plVar7 = (longlong *)
                                                                               FUN_180376210(param_1
                                                  ,plVar6);
                                                  if ((-1 < (int)*plVar6) &&
                                                     (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                    (**(code **)(plVar6[1] + 0x30))(plVar6);
                                                  }
                                                  if (plVar7 == (longlong *)0x0) {
                                                    plVar6 = *(longlong **)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    plVar16 = local_1a0;
                                                    plVar7 = local_1a8;
                                                  }
                                                  else {
                                                    plVar16 = (longlong *)FUN_18037f0c0();
                                                    if (plVar16 == (longlong *)0x0) {
                                                      plVar6 = *(longlong **)(param_1 + 0x60);
                                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                                      if ((-1 < (int)*plVar7) &&
                                                         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                      }
                                                      uVar14 = 0x374;
                                                      plVar16 = local_1a0;
                                                      plVar7 = local_1a8;
                                                    }
                                                    else {
                                                      plVar6 = (longlong *)FUN_18037f0c0();
                                                      if (plVar6 == (longlong *)0x0) {
                                                        plVar6 = *(longlong **)(param_1 + 0x60);
                                                        *(undefined8 *)(param_1 + 0x60) = 0;
                                                        if ((-1 < (int)*plVar7) &&
                                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                        }
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          (**(code **)(plVar16[1] + 0x30))(plVar16);
                                                        }
                                                        uVar14 = 0x375;
                                                        plVar16 = local_1a0;
                                                        plVar7 = local_1a8;
                                                      }
                                                      else {
                                                        local_188 = local_160;
                                                        uVar14 = 0x36f;
                                                        local_180 = local_150;
                                                        local_b8 = local_138;
                                                        *(undefined4 *)(plVar9 + 5) = 0x36f;
                                                        local_228 = (longlong *)DAT_1803b4c10;
                                                        local_b0 = plVar7;
                                                        local_a8 = plVar16;
                                                        local_a0 = plVar6;
                                                        plVar5 = (longlong *)
                                                                 FUN_18037d650(param_1,lVar8,
                                                                               &local_188,&local_b8)
                                                        ;
                                                        if ((-1 < (int)*plVar7) &&
                                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                        }
                                                        if ((-1 < (int)*plVar16) &&
                                                           (*plVar16 = *plVar16 + -1, *plVar16 == 0)
                                                           ) {
                                                          (**(code **)(plVar16[1] + 0x30))(plVar16);
                                                        }
                                                        if ((-1 < (int)*plVar6) &&
                                                           (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                                                          (**(code **)(plVar6[1] + 0x30))(plVar6);
                                                        }
                                                        if (plVar5 == (longlong *)0x0) {
                                                          plVar6 = *(longlong **)(param_1 + 0x60);
                                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                                          plVar16 = local_1a0;
                                                          plVar7 = local_1a8;
                                                        }
                                                        else {
                                                          if ((-1 < (int)*plVar5) &&
                                                             (*plVar5 = *plVar5 + -1, *plVar5 == 0))
                                                          {
                                                            (**(code **)(plVar5[1] + 0x30))(plVar5);
                                                          }
                                                          plVar6 = local_168;
                                                          plVar7 = (longlong *)FUN_18037f0c0();
                                                          if (plVar7 == (longlong *)0x0) {
                                                            plVar6 = *(longlong **)(param_1 + 0x60);
                                                            uVar14 = 0x377;
                                                            *(undefined8 *)(param_1 + 0x60) = 0;
                                                            plVar16 = local_1a0;
                                                            plVar7 = local_1a8;
                                                          }
                                                          else {
                                                            iVar4 = FUN_180001510(plVar7);
                                                            if (iVar4 == -1) {
                                                              plVar6 = *(longlong **)
                                                                        (param_1 + 0x60);
                                                              *(undefined8 *)(param_1 + 0x60) = 0;
                                                              if ((-1 < (int)*plVar7) &&
                                                                 (*plVar7 = *plVar7 + -1,
                                                                 *plVar7 == 0)) {
                                                                (**(code **)(plVar7[1] + 0x30))
                                                                          (plVar7);
                                                              }
                                                              uVar14 = 0x377;
                                                              plVar16 = local_1a0;
                                                              plVar7 = local_1a8;
                                                            }
                                                            else {
                                                              if (iVar4 != 1) {
                                                                if ((-1 < (int)*plVar7) &&
                                                                   (*plVar7 = *plVar7 + -1,
                                                                   *plVar7 == 0)) {
                                                                  (**(code **)(plVar7[1] + 0x30))
                                                                            (plVar7);
                                                                }
                                                                plVar7 = DAT_1803b3f30;
                                                                iVar4 = (int)*DAT_1803b3f30 + 1;
                                                                if (iVar4 != 0) {
                                                                  *(int *)DAT_1803b3f30 = iVar4;
                                                                }
                                                              }
                                                              plVar16 = (longlong *)FUN_18037f0c0();
                                                              if ((-1 < (int)*plVar7) &&
                                                                 (*plVar7 = *plVar7 + -1,
                                                                 *plVar7 == 0)) {
                                                                (**(code **)(plVar7[1] + 0x30))
                                                                          (plVar7);
                                                              }
                                                              uVar14 = 0x377;
                                                              if (plVar16 != (longlong *)0x0) {
                                                                *(undefined4 *)(plVar9 + 5) = 0x377;
                                                                plVar9 = (longlong *)
                                                                         FUN_180376210(param_1,
                                                  plVar16);
                                                  if ((-1 < (int)*plVar16) &&
                                                     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
                                                    (**(code **)(plVar16[1] + 0x30))(plVar16);
                                                  }
                                                  if (plVar9 != (longlong *)0x0) {
                                                    local_118 = plVar9;
                                                    iVar4 = FUN_180001510(plVar9);
                                                    plVar7 = local_1a8;
                                                    if (iVar4 != -1) {
                                                      if (iVar4 == 0) {
                                                        if (local_1a8 == (longlong *)0x0) {
                                                          plVar9 = (longlong *)
                                                                   PyUnicode_FromFormat
                                                                             (
                                                  "cannot access local variable \'%U\' where it is not associated with a value"
                                                  ,DAT_1803b3fb8);
                                                  plVar6 = (longlong *)
                                                           FUN_180376770(param_1,*(undefined8 *)
                                                                                                                                                                    
                                                  PyExc_UnboundLocalError_exref,plVar9);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                  }
                                                  FUN_18000d1d0(param_1,plVar6);
                                                  uVar14 = 0x37b;
                                                  plVar16 = local_1a0;
                                                  plVar9 = local_190;
                                                  }
                                                  else {
                                                    uVar14 = 0x37b;
                                                    *(undefined4 *)(local_190 + 5) = 0x37b;
                                                    plVar16 = (longlong *)
                                                              FUN_18037e6c0(param_1,local_1a8,
                                                                            DAT_1803b4260,
                                                                            *(undefined8 *)
                                                                             (DAT_1803b4c28 + 0x18))
                                                    ;
                                                    if (plVar16 != (longlong *)0x0) {
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        (**(code **)(plVar16[1] + 0x30))(plVar16);
                                                      }
LAB_18010578b:
                                                      FUN_180001d80(param_1);
                                                      lVar8 = *(longlong *)(param_1 + 0x10);
                                                      puVar12 = *(undefined8 **)(lVar8 + 0x41da0);
                                                      if (puVar12 == (undefined8 *)0x0) {
                                                        puVar12 = (undefined8 *)
                                                                  FUN_18035ebd0(PyTuple_Type_exref,2
                                                                               );
                                                      }
                                                      else {
                                                        *(undefined8 *)(lVar8 + 0x41da0) =
                                                             puVar12[3];
                                                        *(int *)(lVar8 + 0x41e3c) =
                                                             *(int *)(lVar8 + 0x41e3c) + -1;
                                                        *puVar12 = 1;
                                                      }
                                                      puVar12[3] = 0;
                                                      puVar12[4] = 0;
                                                      FUN_1801f33c0(puVar12);
                                                      plVar5 = local_90;
                                                      plVar16 = local_c8;
                                                      iVar4 = (int)*local_90 + 1;
                                                      if (iVar4 != 0) {
                                                        *(int *)plVar6 = iVar4;
                                                      }
                                                      puVar12[3] = plVar6;
                                                      iVar4 = (int)*local_c8 + 1;
                                                      if (iVar4 != 0) {
                                                        *(int *)local_c8 = iVar4;
                                                      }
                                                      puVar12[4] = local_c8;
                                                      if ((-1 < (int)*local_e0) &&
                                                         (*local_e0 = *local_e0 + -1, *local_e0 == 0
                                                         )) {
                                                        (**(code **)(local_160[1] + 0x30))();
                                                      }
                                                      if ((-1 < (int)*local_f0) &&
                                                         (*local_f0 = *local_f0 + -1, *local_f0 == 0
                                                         )) {
                                                        (**(code **)(local_148[1] + 0x30))();
                                                      }
                                                      if ((-1 < (int)*local_f8) &&
                                                         (*local_f8 = *local_f8 + -1, *local_f8 == 0
                                                         )) {
                                                        (**(code **)(local_108[1] + 0x30))();
                                                      }
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        (**(code **)(plVar6[1] + 0x30))(plVar6);
                                                      }
                                                      if ((-1 < (int)*local_d8) &&
                                                         (*local_d8 = *local_d8 + -1, *local_d8 == 0
                                                         )) {
                                                        (**(code **)(local_100[1] + 0x30))();
                                                      }
                                                      if ((-1 < (int)*local_d0) &&
                                                         (*local_d0 = *local_d0 + -1, *local_d0 == 0
                                                         )) {
                                                        (**(code **)(local_e8[1] + 0x30))();
                                                      }
                                                      if ((-1 < (int)*local_198) &&
                                                         (*local_198 = *local_198 + -1,
                                                         *local_198 == 0)) {
                                                        (**(code **)(local_120[1] + 0x30))();
                                                      }
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        (**(code **)(plVar16[1] + 0x30))(plVar16);
                                                      }
                                                      if (((local_1a0 != (longlong *)0x0) &&
                                                          (-1 < (int)*local_1a0)) &&
                                                         (*local_1a0 = *local_1a0 + -1,
                                                         *local_1a0 == 0)) {
                                                        (**(code **)(local_1a0[1] + 0x30))();
                                                      }
                                                      if ((-1 < (int)*plVar9) &&
                                                         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                        (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                      }
                                                      if ((-1 < (int)*plVar7) &&
                                                         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                        (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                      }
                                                      if ((-1 < (int)*local_140) &&
                                                         (*local_140 = *local_140 + -1,
                                                         *local_140 == 0)) {
                                                        (**(code **)(local_140[1] + 0x30))();
                                                      }
                                                      if ((-1 < (int)*local_110) &&
                                                         (*local_110 = *local_110 + -1,
                                                         *local_110 == 0)) {
                                                        (**(code **)(local_110[1] + 0x30))();
                                                      }
                                                      if ((-1 < (int)*local_150) &&
                                                         (*local_150 = *local_150 + -1,
                                                         *local_150 == 0)) {
                                                        (**(code **)(local_150[1] + 0x30))();
                                                      }
                                                      if ((-1 < (int)*local_138) &&
                                                         (*local_138 = *local_138 + -1,
                                                         *local_138 == 0)) {
                                                        (**(code **)(local_138[1] + 0x30))();
                                                      }
                                                      if ((-1 < (int)*local_130) &&
                                                         (*local_130 = *local_130 + -1,
                                                         *local_130 == 0)) {
                                                        (**(code **)(local_130[1] + 0x30))();
                                                      }
                                                      if ((int)*local_128 < 0) {
                                                        return puVar12;
                                                      }
                                                      *local_128 = *local_128 + -1;
                                                      if (*local_128 != 0) {
                                                        return puVar12;
                                                      }
                                                      (**(code **)(local_128[1] + 0x30))();
                                                      return puVar12;
                                                    }
                                                    plVar6 = *(longlong **)(param_1 + 0x60);
                                                    *(undefined8 *)(param_1 + 0x60) = 0;
                                                    plVar16 = local_1a0;
                                                    plVar9 = local_190;
                                                  }
                                                  }
                                                  else if (local_1a8 == (longlong *)0x0) {
                                                    plVar9 = (longlong *)
                                                             PyUnicode_FromFormat
                                                                       (
                                                  "cannot access local variable \'%U\' where it is not associated with a value"
                                                  ,DAT_1803b3fb8);
                                                  plVar6 = (longlong *)
                                                           FUN_180376770(param_1,*(undefined8 *)
                                                                                                                                                                    
                                                  PyExc_UnboundLocalError_exref,plVar9);
                                                  if ((-1 < (int)*plVar9) &&
                                                     (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
                                                    (**(code **)(plVar9[1] + 0x30))(plVar9);
                                                  }
                                                  FUN_18000d1d0(param_1,plVar6);
                                                  uVar14 = 0x379;
                                                  plVar16 = local_1a0;
                                                  plVar9 = local_190;
                                                  }
                                                  else {
                                                    plVar16 = (longlong *)FUN_18037f0c0();
                                                    piVar1 = DAT_1803b4c18;
                                                    if (plVar16 == (longlong *)0x0) {
                                                      plVar6 = *(longlong **)(param_1 + 0x60);
                                                      uVar14 = 0x379;
                                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                                      plVar16 = local_1a0;
                                                      plVar9 = local_190;
                                                    }
                                                    else {
                                                      lVar8 = *(longlong *)(param_1 + 0x10);
                                                      plVar5 = *(longlong **)(lVar8 + 0x41da0);
                                                      if (plVar5 == (longlong *)0x0) {
                                                        plVar5 = (longlong *)
                                                                 FUN_18035ebd0(PyTuple_Type_exref,2)
                                                        ;
                                                      }
                                                      else {
                                                        *(longlong *)(lVar8 + 0x41da0) = plVar5[3];
                                                        *(int *)(lVar8 + 0x41e3c) =
                                                             *(int *)(lVar8 + 0x41e3c) + -1;
                                                        *plVar5 = 1;
                                                      }
                                                      plVar5[3] = 0;
                                                      plVar5[4] = 0;
                                                      FUN_1801f33c0(plVar5);
                                                      if (*piVar1 + 1 != 0) {
                                                        *piVar1 = *piVar1 + 1;
                                                      }
                                                      plVar5[3] = (longlong)piVar1;
                                                      plVar7 = (longlong *)
                                                               FUN_180024e50(param_1,plVar9,
                                                                             DAT_1803b4c20);
                                                      if (plVar7 != (longlong *)0x0) {
                                                        lVar8 = FUN_180375520(param_1,plVar7,
                                                                              DAT_1803b3f30);
                                                        if ((-1 < (int)*plVar7) &&
                                                           (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                                          (**(code **)(plVar7[1] + 0x30))(plVar7);
                                                        }
                                                        if (lVar8 != 0) {
                                                          plVar5[4] = lVar8;
                                                          plVar7 = (longlong *)
                                                                   PyUnicode_Join(DAT_1803b3f30,
                                                                                  plVar5);
                                                          if ((-1 < (int)*plVar5) &&
                                                             (*plVar5 = *plVar5 + -1, *plVar5 == 0))
                                                          {
                                                            (**(code **)(plVar5[1] + 0x30))(plVar5);
                                                          }
                                                          if (plVar7 != (longlong *)0x0) {
                                                            uVar14 = 0x379;
                                                            *(undefined4 *)(local_190 + 5) = 0x379;
                                                            plVar5 = (longlong *)
                                                                     FUN_180376770(param_1,plVar16,
                                                                                   plVar7);
                                                            if ((-1 < (int)*plVar16) &&
                                                               (*plVar16 = *plVar16 + -1,
                                                               *plVar16 == 0)) {
                                                              (**(code **)(plVar16[1] + 0x30))
                                                                        (plVar16);
                                                            }
                                                            if ((-1 < (int)*plVar7) &&
                                                               (*plVar7 = *plVar7 + -1, *plVar7 == 0
                                                               )) {
                                                              (**(code **)(plVar7[1] + 0x30))
                                                                        (plVar7);
                                                            }
                                                            if (plVar5 != (longlong *)0x0) {
                                                              plVar6 = local_168;
                                                              plVar7 = local_1a8;
                                                              if ((-1 < (int)*plVar5) &&
                                                                 (*plVar5 = *plVar5 + -1,
                                                                 *plVar5 == 0)) {
                                                                (**(code **)(plVar5[1] + 0x30))
                                                                          (plVar5);
                                                                plVar6 = local_168;
                                                                plVar7 = local_1a8;
                                                              }
                                                              goto LAB_18010578b;
                                                            }
                                                            goto LAB_1801044f7;
                                                          }
                                                          plVar6 = *(longlong **)(param_1 + 0x60);
                                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                                          if ((-1 < (int)*plVar16) &&
                                                             (*plVar16 = *plVar16 + -1,
                                                             *plVar16 == 0)) {
                                                            (**(code **)(plVar16[1] + 0x30))
                                                                      (plVar16);
                                                          }
                                                          uVar14 = 0x379;
                                                          plVar16 = local_1a0;
                                                          plVar9 = local_190;
                                                          plVar7 = local_1a8;
                                                          goto LAB_18010450f;
                                                        }
                                                      }
                                                      uVar14 = 0x379;
                                                      plVar6 = *(longlong **)(param_1 + 0x60);
                                                      *(undefined8 *)(param_1 + 0x60) = 0;
                                                      if ((-1 < (int)*plVar16) &&
                                                         (*plVar16 = *plVar16 + -1, *plVar16 == 0))
                                                      {
                                                        (**(code **)(plVar16[1] + 0x30))(plVar16);
                                                      }
                                                      plVar16 = local_1a0;
                                                      plVar9 = local_190;
                                                      plVar7 = local_1a8;
                                                      if ((-1 < (int)*plVar5) &&
                                                         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                                        (**(code **)(plVar5[1] + 0x30))(plVar5);
                                                        plVar16 = local_1a0;
                                                        plVar9 = local_190;
                                                        plVar7 = local_1a8;
                                                      }
                                                    }
                                                  }
                                                  goto LAB_18010450f;
                                                  }
                                                  uVar14 = 0x378;
                                                  }
                                                  goto LAB_1801044f7;
                                                  }
                                                  plVar6 = *(longlong **)(param_1 + 0x60);
                                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                                  plVar16 = local_1a0;
                                                  plVar7 = local_1a8;
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
                                                  }
                                                  }
                                                  }
                                                  }
                                                  goto LAB_18010450f;
                                                  }
                                                }
                                                plVar6 = *(longlong **)(param_1 + 0x60);
                                                uVar14 = 0x367;
                                                *(undefined8 *)(param_1 + 0x60) = 0;
                                                plVar9 = local_190;
                                                plVar7 = local_1a8;
                                                goto LAB_18010450f;
                                              }
                                            }
                                            uVar14 = 0x365;
                                            goto LAB_1801044f7;
                                          }
                                          plVar6 = *(longlong **)(param_1 + 0x60);
                                          *(undefined8 *)(param_1 + 0x60) = 0;
                                          if ((-1 < (int)*plVar7) &&
                                             (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
                                            (**(code **)(plVar7[1] + 0x30))(plVar7);
                                          }
                                          if ((-1 < (int)*plVar5) &&
                                             (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
                                            (**(code **)(plVar5[1] + 0x30))(plVar5);
                                          }
                                          uVar14 = 0x361;
                                          plVar16 = local_1a0;
                                          plVar9 = local_190;
                                          plVar7 = local_1a8;
                                        }
                                        goto LAB_18010450f;
                                      }
                                      uVar14 = 0x35c;
                                    }
                                    goto LAB_1801044f7;
                                  }
                                  plVar6 = *(longlong **)(param_1 + 0x60);
                                  *(undefined8 *)(param_1 + 0x60) = 0;
                                  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0))
                                  {
                                    (**(code **)(plVar7[1] + 0x30))(plVar7);
                                  }
                                  uVar14 = 0x35b;
                                  plVar16 = local_1a0;
                                  plVar7 = local_1a8;
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                    goto LAB_18010450f;
                  }
                }
LAB_1801044f7:
                plVar6 = *(longlong **)(param_1 + 0x60);
                *(undefined8 *)(param_1 + 0x60) = 0;
                plVar16 = local_1a0;
                plVar9 = local_190;
                plVar7 = local_1a8;
              }
            }
          }
        }
      }
    }
  }
LAB_18010450f:
  piVar1 = (int *)plVar6[5];
  if (piVar1 == (int *)0x0) {
    pcVar10 = (code *)FUN_18038e2a0(plVar9,uVar14);
    plVar5 = (longlong *)plVar6[5];
    pcVar11 = pcVar10;
    if (pcVar10 == _Py_NoneStruct_exref) {
      pcVar11 = (code *)0x0;
    }
    else {
LAB_1801058aa:
      pcVar10 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar1 + 6) == plVar9) goto LAB_1801058f3;
    pcVar10 = (code *)FUN_18038e2a0(plVar9,uVar14);
    *(int **)(pcVar10 + 0x10) = piVar1;
    if (*piVar1 + 1 != 0) {
      *piVar1 = *piVar1 + 1;
    }
    plVar5 = (longlong *)plVar6[5];
    pcVar11 = pcVar10;
    if (pcVar10 != _Py_NoneStruct_exref) goto LAB_1801058aa;
    pcVar11 = (code *)0x0;
  }
  plVar6[5] = (longlong)pcVar11;
  if (((plVar5 != (longlong *)0x0) && (-1 < (int)*plVar5)) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)
     ) {
    (**(code **)(plVar5[1] + 0x30))();
  }
  if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
  }
LAB_1801058f3:
  plVar2 = local_168;
  local_1b8 = local_118;
  local_1c8 = local_158;
  local_1d0 = local_120;
  local_1d8 = local_e8;
  local_1e0 = local_100;
  local_1e8 = local_168;
  local_1f0 = local_108;
  local_1f8 = local_148;
  local_200 = local_160;
  local_208 = local_128;
  local_210 = local_130;
  local_218 = local_138;
  local_220 = local_150;
  local_228 = local_110;
  local_1c0 = plVar16;
  FUN_18036d0a0(plVar9,"ooooooooooooooooo",plVar7,local_140);
  plVar5 = DAT_1803b3958;
  if (plVar9 == DAT_1803b3958) {
    if ((-1 < (int)*DAT_1803b3958) && (*DAT_1803b3958 = *DAT_1803b3958 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))();
    }
    DAT_1803b3958 = (longlong *)0x0;
  }
  lVar8 = **(longlong **)(param_1 + 0x38);
  plVar9 = *(longlong **)(lVar8 + 0x30);
  **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar8 + 8);
  *(undefined4 *)(plVar9 + 8) = 0xffffffff;
  plVar5 = (longlong *)plVar9[2];
  if (((plVar5 != (longlong *)0x0) && (plVar9[2] = 0, -1 < (int)*plVar5)) &&
     (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
    (**(code **)(plVar5[1] + 0x30))();
  }
  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  plVar9[10] = 0;
  if (((local_160 != (longlong *)0x0) && (-1 < (int)*local_160)) &&
     (*local_160 = *local_160 + -1, *local_160 == 0)) {
    (**(code **)(local_160[1] + 0x30))();
  }
  if (((local_148 != (longlong *)0x0) && (-1 < (int)*local_148)) &&
     (*local_148 = *local_148 + -1, *local_148 == 0)) {
    (**(code **)(local_148[1] + 0x30))();
  }
  if (((local_108 != (longlong *)0x0) && (-1 < (int)*local_108)) &&
     (*local_108 = *local_108 + -1, *local_108 == 0)) {
    (**(code **)(local_108[1] + 0x30))();
  }
  if (((plVar2 != (longlong *)0x0) && (-1 < (int)*plVar2)) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)
     ) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  if (((local_100 != (longlong *)0x0) && (-1 < (int)*local_100)) &&
     (*local_100 = *local_100 + -1, *local_100 == 0)) {
    (**(code **)(local_100[1] + 0x30))();
  }
  if (((local_e8 != (longlong *)0x0) && (-1 < (int)*local_e8)) &&
     (*local_e8 = *local_e8 + -1, *local_e8 == 0)) {
    (**(code **)(local_e8[1] + 0x30))();
  }
  if (((local_120 != (longlong *)0x0) && (-1 < (int)*local_120)) &&
     (*local_120 = *local_120 + -1, *local_120 == 0)) {
    (**(code **)(local_120[1] + 0x30))();
  }
  if (((local_158 != (longlong *)0x0) && (-1 < (int)*local_158)) &&
     (*local_158 = *local_158 + -1, *local_158 == 0)) {
    (**(code **)(local_158[1] + 0x30))();
  }
  if (((plVar16 != (longlong *)0x0) && (-1 < (int)*plVar16)) &&
     (*plVar16 = *plVar16 + -1, *plVar16 == 0)) {
    (**(code **)(plVar16[1] + 0x30))(plVar16);
  }
  if (((local_118 != (longlong *)0x0) && (-1 < (int)*local_118)) &&
     (*local_118 = *local_118 + -1, *local_118 == 0)) {
    (**(code **)(local_118[1] + 0x30))();
  }
  if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if ((-1 < (int)*local_140) && (*local_140 = *local_140 + -1, *local_140 == 0)) {
    (**(code **)(local_140[1] + 0x30))();
  }
  if ((-1 < (int)*local_110) && (*local_110 = *local_110 + -1, *local_110 == 0)) {
    (**(code **)(local_110[1] + 0x30))();
  }
  if ((-1 < (int)*local_150) && (*local_150 = *local_150 + -1, *local_150 == 0)) {
    (**(code **)(local_150[1] + 0x30))();
  }
  if ((-1 < (int)*local_138) && (*local_138 = *local_138 + -1, *local_138 == 0)) {
    (**(code **)(local_138[1] + 0x30))();
  }
  if ((-1 < (int)*local_130) && (*local_130 = *local_130 + -1, *local_130 == 0)) {
    (**(code **)(local_130[1] + 0x30))();
  }
  if ((-1 < (int)*local_128) && (*local_128 = *local_128 + -1, *local_128 == 0)) {
    (**(code **)(local_128[1] + 0x30))();
  }
  plVar7 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar6;
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))();
  }
  return (undefined8 *)0x0;
}

